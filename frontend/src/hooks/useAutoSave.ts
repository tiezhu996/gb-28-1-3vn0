// 答卷自动保存 hook：作答变化后短暂停顿（debounce）写入服务端，支持断点续答。
// 规则：
// 1. 每次保存携带单调递增的 save_version，服务端只接受比已存版本更新的请求（乱序旧请求被忽略）；
// 2. 保存失败只提示，不改页面答案、不改已保存记录，可手动重试；
// 3. 若服务端返回的版本号小于本次发送版本，说明被其他设备的更新抢先，进入冲突态，需刷新页面；
// 4. 页面隐藏/卸载时用 keepalive 请求尽力保存最后一次未落盘的草稿。
'use client';
import { useCallback, useEffect, useRef, useState } from 'react';
import { recordApi, type AnswerInput } from '@/api/record';
import { getAuthToken } from '@/utils/request';
import type { ExamRecord } from '@/types';

export type SaveStatus = 'idle' | 'saving' | 'saved' | 'error' | 'conflict';

const AUTOSAVE_DELAY = 1500;
const API_BASE = '/api/v1';

interface PendingDraft {
  answers: AnswerInput[];
  currentIndex: number;
  version: number;
}

export function useAutoSave(record: ExamRecord | null) {
  const [status, setStatus] = useState<SaveStatus>('idle');
  const [lastSavedAt, setLastSavedAt] = useState<string | null>(null);

  const recordIdRef = useRef('');
  const versionRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingRef = useRef<PendingDraft | null>(null);
  const conflictRef = useRef(false);
  const inFlightRef = useRef(false);

  // 首次拿到记录时，以服务端版本号与最后保存时间为基线（刷新/换设备续答）
  useEffect(() => {
    if (record && recordIdRef.current === '') {
      recordIdRef.current = record.id;
      versionRef.current = record.save_version ?? 0;
      setLastSavedAt(record.last_saved_at ?? null);
      if (record.last_saved_at) setStatus('saved');
    }
  }, [record]);

  const doSave = useCallback(async (keepalive = false): Promise<boolean> => {
    const payload = pendingRef.current;
    const recordId = recordIdRef.current;
    if (!payload || !recordId || conflictRef.current) {
      // 没有待保存内容：视为成功；冲突态交页面提示用户刷新
      return payload === null;
    }
    // 普通保存同一时刻只允许一个在飞请求，防止并发乱序；keepalive 只在页面隐藏时触发且不读响应
    if (!keepalive && inFlightRef.current) return false;
    inFlightRef.current = true;
    setStatus('saving');
    try {
      let rec: ExamRecord;
      if (keepalive) {
        // 页面卸载/隐藏：fetch keepalive 尽力送达，不处理响应
        const res = await fetch(`${API_BASE}/exam-records/${recordId}/autosave`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json', ...(getAuthToken() ? { Authorization: `Bearer ${getAuthToken()}` } : {}) },
          body: JSON.stringify({ answers: payload.answers, current_index: payload.currentIndex, save_version: payload.version }),
          keepalive: true,
        });
        inFlightRef.current = false;
        return res.ok;
      }
      rec = await recordApi.autoSave(recordId, {
        answers: payload.answers,
        current_index: payload.currentIndex,
        save_version: payload.version,
      });
      if (rec.save_version < payload.version) {
        // 服务端忽略了本次请求：已有来自其他设备/标签页的更新版本，提示刷新，避免覆盖新答案
        conflictRef.current = true;
        setStatus('conflict');
        return false;
      }
      versionRef.current = rec.save_version;
      setLastSavedAt(rec.last_saved_at ?? null);
      // 保存飞行期间又产生了更新的草稿：继续保存更新版本
      if (pendingRef.current && pendingRef.current.version > payload.version) {
        inFlightRef.current = false;
        return doSave(false);
      }
      pendingRef.current = null;
      setStatus('saved');
      return true;
    } catch {
      // 保存失败：保留 pending 与页面答案不变，等待重试或下次变化后再次尝试
      setStatus('error');
      return false;
    } finally {
      inFlightRef.current = false;
    }
  }, []);

  // 作答/题号变化后短暂停顿再保存
  const scheduleSave = useCallback((answers: AnswerInput[], currentIndex: number) => {
    if (!recordIdRef.current || conflictRef.current) return;
    if (timerRef.current) clearTimeout(timerRef.current);
    versionRef.current += 1;
    pendingRef.current = { answers, currentIndex, version: versionRef.current };
    setStatus('saving');
    timerRef.current = setTimeout(() => {
      void doSave(false);
    }, AUTOSAVE_DELAY);
  }, [doSave]);

  // 立即保存待写内容（交卷前调用）；无待写内容时直接返回成功
  const flush = useCallback(async () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    if (!pendingRef.current) return true;
    return doSave(false);
  }, [doSave]);

  // 失败后手动重试（沿用最后一次待保存内容，分配新版本号）
  const retry = useCallback(() => {
    if (!pendingRef.current || conflictRef.current) return;
    versionRef.current += 1;
    pendingRef.current = { ...pendingRef.current, version: versionRef.current };
    void doSave(false);
  }, [doSave]);

  // 切后台 / 关闭页面前尽力把未落盘草稿发出
  useEffect(() => {
    const onHidden = () => {
      if (document.hidden && pendingRef.current && !conflictRef.current) {
        if (timerRef.current) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
        versionRef.current += 1;
        pendingRef.current = { ...pendingRef.current, version: versionRef.current };
        void doSave(true);
      }
    };
    const onPageHide = () => onHidden();
    document.addEventListener('visibilitychange', onHidden);
    window.addEventListener('pagehide', onPageHide);
    return () => {
      document.removeEventListener('visibilitychange', onHidden);
      window.removeEventListener('pagehide', onPageHide);
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [doSave]);

  return { status, lastSavedAt, scheduleSave, flush, retry };
}
