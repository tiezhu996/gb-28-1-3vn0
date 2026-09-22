'use client';
import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { useQuestionStore } from '@/stores/questionStore';
import { DataTable, type Column } from '@/components/DataTable';
import { Pagination } from '@/components/Pagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { DifficultyBadge, QuestionTypeBadge, StatusBadge } from '@/components/StatusBadge';
import { questionApi } from '@/api/question';
import { difficultyText, formatDateTime } from '@/utils/format';
import { DIFFICULTY, QUESTION_STATUS, SUBJECTS } from '@/constants';
import type { Question } from '@/types';

export default function QuestionsPage() {
  const { isTeacher, isAdmin } = useAuth();
  const canWrite = isTeacher || isAdmin;
  const { list, total, loading, fetch, remove } = useQuestionStore();
  const [page, setPage] = useState(1);
  const [subject, setSubject] = useState('');
  const [type, setType] = useState('');
  const [difficulty, setDifficulty] = useState('');
  const [keyword, setKeyword] = useState('');
  const [confirmId, setConfirmId] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const reload = useCallback(() => {
    fetch({ subject, type, difficulty, keyword, page, page_size: 10 });
  }, [fetch, subject, type, difficulty, keyword, page]);

  useEffect(() => {
    reload();
  }, [reload]);

  const columns: Column<Question>[] = [
    { key: 'content', title: '题干', render: (q) => <span className="line-clamp-1 max-w-xs">{q.content}</span> },
    { key: 'type', title: '题型', render: (q) => <QuestionTypeBadge type={q.type} /> },
    { key: 'subject', title: '学科', render: (q) => <span className="text-xs text-gray-500">{q.subject}</span> },
    { key: 'difficulty', title: '难度', render: (q) => <DifficultyBadge difficulty={q.difficulty} /> },
    { key: 'knowledge_points', title: '知识点', render: (q) => (
        <span className="text-xs text-gray-500">{(q.knowledge_points ?? []).slice(0, 2).join('、')}</span>
      ) },
    { key: 'score', title: '分值', render: (q) => <span>{q.score}</span> },
    { key: 'status', title: '状态', render: (q) => (
        <StatusBadge text={q.status === QUESTION_STATUS.PUBLISHED ? '已发布' : '草稿'} color={q.status === QUESTION_STATUS.PUBLISHED ? 'green' : 'gray'} />
      ) },
    { key: 'created_at', title: '创建时间', render: (q) => <span className="text-xs">{formatDateTime(q.created_at)}</span> },
    ...(canWrite
      ? [{
          key: 'actions',
          title: '操作',
          render: (q: Question) => (
            <div className="flex gap-2">
              <Link href={`/questions/new?id=${q.id}`} className="text-brand-600 hover:underline">编辑</Link>
              <button onClick={() => setConfirmId(q.id)} className="text-red-600 hover:underline">删除</button>
            </div>
          ),
        } as Column<Question>]
      : []),
  ];

  const onImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const count = await useQuestionStore.getState().importExcel(file);
      alert(`导入成功：${count} 道题`);
      reload();
    } catch (err) {
      alert((err as Error).message);
    } finally {
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-gray-800">题库管理</h1>
        <div className="flex flex-wrap gap-2">
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索题干…"
            className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-brand-500 focus:outline-none" />
          <select value={subject} onChange={(e) => setSubject(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部学科</option>
            {SUBJECTS.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
          <select value={type} onChange={(e) => setType(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部题型</option>
            <option value="single">单选题</option>
            <option value="multiple">多选题</option>
            <option value="judge">判断题</option>
            <option value="fill">填空题</option>
            <option value="short">简答题</option>
          </select>
          <select value={difficulty} onChange={(e) => setDifficulty(e.target.value)}
            className="rounded-lg border border-gray-300 px-2 py-1.5 text-sm">
            <option value="">全部难度</option>
            {Object.values(DIFFICULTY).map((d) => <option key={d} value={d}>{difficultyText(d)}</option>)}
          </select>
          <button onClick={() => { setPage(1); }} className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50">查询</button>
          {canWrite && (
            <>
              <Link href="/questions/new" className="rounded-lg bg-brand-600 px-3 py-1.5 text-sm text-white hover:bg-brand-700">新增题目</Link>
              <button onClick={() => fileRef.current?.click()}
                className="rounded-lg border border-brand-600 px-3 py-1.5 text-sm text-brand-600 hover:bg-brand-50">Excel 导入</button>
              <a href={questionApi.templateUrl()} className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm hover:bg-gray-50">下载模板</a>
              <input ref={fileRef} type="file" accept=".xlsx" className="hidden" onChange={onImport} />
            </>
          )}
        </div>
      </div>

      <DataTable columns={columns} rows={list} loading={loading} emptyTitle="题库为空，快去录入第一道题吧" />
      <Pagination page={page} pageSize={10} total={total} onChange={setPage} />

      <ConfirmDialog
        open={!!confirmId}
        message="确定删除该题目？删除后不可恢复。"
        onCancel={() => setConfirmId(null)}
        onConfirm={async () => {
          if (confirmId) {
            await remove(confirmId);
            setConfirmId(null);
            reload();
          }
        }}
      />
    </div>
  );
}
