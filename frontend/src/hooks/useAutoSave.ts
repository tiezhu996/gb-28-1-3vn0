// 在线答题自动保存（断点续答）。
// 作答/题号变化后停顿 DEBOUNCE_MS 静默写入草稿；基于服务端 answer_version 乐观锁，
// 乱序到达的旧请求会被服务端拒绝（code=5005），此时停止后续保存并提示刷新页面。
// 保存失败只提示，不改动页面答案，也不推进版本号（已保存记录不受影响）。
'use client';
import { useCallback, useEffect, useRef, useState } from 'react';
import { recordApi, type AnswerInput } from '@/api/record';
import { ApiError } from '@/utils/request';

export type AutoSaveStatus = 'idle' | 'pending' | 'saving' | 'saved' | 'error' | 'conflict';

// 服务端错误码：草稿版本过旧（见 backend constants.CodeRecordVersion）
export const CODE_RECORD_VERSION = 5005;
// 服务端错误码：考试记录已超时（截止后不再保存）
export const CODE_RECORD_EXPIRED = 5003;

const DEBOUNCE_MS = 1200;

interface UseAutoSaveInput {
  // recordId 为 null 表示答卷尚未加载完成，此时不发起任何保存
  recordId: string | null;
  answers: Record<string, string>;
  current: number;
  // 从服务端恢复的草稿版本（record.answer_version），recordId 变化时重新同步
  version: number;
}

interface AutoSaveState {
  status: AutoSaveStatus;
  errorMsg: string;
  lastSavedAt: number | null;
  // flush 立即保存一次（交卷前/页面隐藏/离开页面前），返回是否与服务端一致
  flush: () => Promise<boolean>;
  retry: () => void;
}

function signature(answers: Record<string, string>, current: number) {
  return JSON.stringify({ a: answers, c: current });
}

export function useAutoSave({ recordId, answers, current, version }: UseAutoSaveInput): AutoSaveState {
  const [status, setStatus] = useState<AutoSaveStatus>('idle');
  const [errorMsg, setErrorMsg] = useState('');
  const [lastSavedAt, setLastSavedAt] = useState<number | null>(null);

  // 最新输入（供定时器/回调读取，避免闭包过期）
  const stateRef = useRef({ recordId, answers, current });
  stateRef.current = { recordId, answers, current };

  const versionRef = useRef(version);
  const conflictRef = useRef(false);
  // 已与服务端达成一致的内容签名；用于初始化去抖与识别待保存改动
  const syncedSigRef = useRef('');
  const initializedRef = useRef(false);
  const inflightRef = useRef<Promise<boolean> | null>(null);
  const catchupRef = useRef(false);

  // 切换答卷（重新进入/换一份记录）时重置全部状态
  useEffect(() => {
    versionRef.current = version;
    conflictRef.current = false;
    catchupRef.current = false;
    initializedRef.current = false;
    syncedSigRef.current = '';
    setStatus('idle');
    setErrorMsg('');
    setLastSavedAt(null);
  }, [recordId, version]);

  const runSave = useCallback(async (keepalive: boolean): Promise<boolean> => {
    const { recordId: id, answers: ans, current: cur } = stateRef.current;
    if (!id || conflictRef.current) return false;
    const sentSig = signature(ans, cur);
    if (sentSig === syncedSigRef.current) return true;

    const ansList: AnswerInput[] = Object.entries(ans).map(([question_id, answer]) => ({ question_id, answer }));
    setStatus('saving');
    try {
      const res = await recordApi.saveDraft(id, ansList, cur, versionRef.current, keepalive);
      // 成功后才推进本地版本与“已保存”基线；失败时两者都不动，页面答案与服务端记录均不变
      versionRef.current = res.version;
      syncedSigRef.current = sentSig;
      setLastSavedAt(Date.now());
      setErrorMsg('');
      setStatus('saved');
      return true;
    } catch (err) {
      if (err instanceof ApiError && err.code === CODE_RECORD_VERSION) {
        // 服务端已有更新的草稿（通常是另一设备作答）：本地继续保存只会被忽略，停止并提示刷新
        conflictRef.current = true;
        setStatus('conflict');
        setErrorMsg('检测到该答卷在其他页面或设备上有更新的作答，本地内容已停止自动保存。请刷新页面加载最新答案后再继续。');
      } else {
        setStatus('error');
        setErrorMsg(err instanceof ApiError && err.code === CODE_RECORD_EXPIRED ? '考试已截止，答案未能保存' : `答案自动保存失败：${(err as Error).message}`);
      }
      return false;
    }
  }, []);

  // runRef：保证“飞行中再次触发”能在结束后补存最新内容
  const runRef = useRef<(keepalive: boolean) => Promise<boolean>>(runSave);
  runRef.current = runSave;

  const trigger = useCallback((keepalive: boolean) => {
    if (inflightRef.current) {
      catchupRef.current = true;
      return;
    }
    const p = runRef.current(keepalive);
    inflightRef.current = p;
    void p.finally(() => {
      inflightRef.current = null;
      if (catchupRef.current) {
        catchupRef.current = false;
        const { answers: ans, current: cur } = stateRef.current;
        if (!conflictRef.current && signature(ans, cur) !== syncedSigRef.current) {
          trigger(false);
        }
      }
    });
  }, []);

  // 作答/题号变化后短暂停顿再写入
  useEffect(() => {
    if (!recordId) return;
    if (!initializedRef.current) {
      // 首次为从服务端恢复的内容，跳过保存
      initializedRef.current = true;
      syncedSigRef.current = signature(answers, current);
      return;
    }
    if (conflictRef.current) return;
    if (signature(answers, current) === syncedSigRef.current) return;
    setStatus('pending');
    const timer = window.setTimeout(() => trigger(false), DEBOUNCE_MS);
    return () => window.clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [recordId, answers, current]);

  const flush = useCallback(async () => {
    if (!stateRef.current.recordId || conflictRef.current) return false;
    const { answers: ans, current: cur } = stateRef.current;
    if (signature(ans, cur) === syncedSigRef.current) return true;
    if (inflightRef.current) {
      await inflightRef.current;
      if (signature(stateRef.current.answers, stateRef.current.current) === syncedSigRef.current) return true;
    }
    return runRef.current(false);
  }, []);

  // 切到后台（不卸载）时尽快落盘一次
  useEffect(() => {
    if (!recordId) return;
    const onHidden = () => {
      if (document.visibilityState === 'hidden') void flush();
    };
    // pagehide（刷新/关闭/跳转）：keepalive 请求允许在页面卸载后继续发出。
    // 即便已有飞行中的保存请求，也再发一个 keepalive 请求兜底（两者内容一致，服务端版本锁保证只落盘一份）。
    const onPageHide = () => {
      const { answers: ans, current: cur } = stateRef.current;
      if (!conflictRef.current && signature(ans, cur) !== syncedSigRef.current) {
        void runRef.current(true);
      }
    };
    const onOnline = () => {
      const { answers: ans, current: cur } = stateRef.current;
      if (!conflictRef.current && signature(ans, cur) !== syncedSigRef.current) trigger(false);
    };
    document.addEventListener('visibilitychange', onHidden);
    window.addEventListener('pagehide', onPageHide);
    window.addEventListener('online', onOnline);
    return () => {
      document.removeEventListener('visibilitychange', onHidden);
      window.removeEventListener('pagehide', onPageHide);
      window.removeEventListener('online', onOnline);
    };
  }, [recordId, flush, trigger]);

  const retry = useCallback(() => {
    const { answers: ans, current: cur } = stateRef.current;
    if (conflictRef.current || signature(ans, cur) === syncedSigRef.current) return;
    trigger(false);
  }, [trigger]);

  return { status, errorMsg, lastSavedAt, flush, retry };
}
