// 考试倒计时 hook：剩余秒数、格式化、结束回调。
// 说明：endAt <= 0 表示尚未就绪（记录未加载），不触发结束；仅当真实时间越过 endAt 才触发 onEnd。
'use client';
import { useEffect, useMemo, useState } from 'react';

export function useCountdown(endAt: number, onEnd?: () => void) {
  const [remaining, setRemaining] = useState(() =>
    endAt > 0 ? Math.max(0, Math.floor((endAt - Date.now()) / 1000)) : 0,
  );

  useEffect(() => {
    if (endAt <= 0) {
      setRemaining(0);
      return;
    }
    setRemaining(Math.max(0, Math.floor((endAt - Date.now()) / 1000)));
    const timer = window.setInterval(() => {
      setRemaining((r) => Math.max(0, r - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [endAt]);

  useEffect(() => {
    if (endAt > 0 && remaining <= 0 && Date.now() >= endAt) {
      onEnd?.();
    }
  }, [remaining, endAt, onEnd]);

  const text = useMemo(() => {
    const h = Math.floor(remaining / 3600);
    const m = Math.floor((remaining % 3600) / 60);
    const s = remaining % 60;
    const p = (n: number) => String(n).padStart(2, '0');
    return `${p(h)}:${p(m)}:${p(s)}`;
  }, [remaining]);

  const warn5 = remaining <= 300 && remaining > 0;

  return { remaining, text, warn5 };
}
