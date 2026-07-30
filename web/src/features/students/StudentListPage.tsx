import { type FormEvent, useEffect, useState } from "react";

import { api, ApiClientError } from "../../api/client";
import type { ClassInfo, Student, StudentPage } from "../../api/types";

type Props = {
  currentClass: ClassInfo;
  initialPage: StudentPage | null;
  openCreateInitially?: boolean;
  onChanged: (page: StudentPage) => void;
};

export function StudentListPage({ currentClass, initialPage, openCreateInitially, onChanged }: Props) {
  const [page, setPage] = useState(initialPage);
  const [search, setSearch] = useState("");
  const [showForm, setShowForm] = useState(Boolean(openCreateInitially));
  const [editing, setEditing] = useState<Student | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!page) void load("");
  }, []);

  async function load(term = search, includeTrashed = false) {
    try {
      const result = await api.get<StudentPage>(
        `/students?classId=${currentClass.id}&page=1&pageSize=50&search=${encodeURIComponent(term)}${includeTrashed ? "&includeTrashed=true" : ""}`
      );
      setPage(result);
      onChanged(result);
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "加载学生名单失败");
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    const form = new FormData(event.currentTarget);
    const body = {
      classId: currentClass.id,
      name: form.get("name"),
      gender: form.get("gender"),
      studentNumber: form.get("studentNumber"),
      status: "active",
      notes: form.get("notes")
    };
    try {
      if (editing) {
        await api.patch<Student>(`/students/${editing.id}`, body);
      } else {
        await api.post<Student>("/students", body);
      }
      setShowForm(false);
      setEditing(null);
      await load();
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "保存学生失败");
    }
  }

  async function trash(student: Student) {
    if (!window.confirm(`确认将“${student.name}”移入回收站？`)) return;
    await api.post(`/students/${student.id}/trash`);
    await load();
  }

  return (
    <>
      <header className="page-header">
        <div>
          <p className="eyebrow">STUDENT DIRECTORY · 花名册</p>
          <h2>学生名单</h2>
          <p>共 {page?.total ?? 0} 名学生，数据来自 {currentClass.name}。</p>
        </div>
        <button className="primary-button compact" onClick={() => { setEditing(null); setShowForm(true); }} type="button">
          ＋ 添加学生
        </button>
      </header>

      <section className="toolbar">
        <form onSubmit={(event) => { event.preventDefault(); void load(); }}>
          <input
            aria-label="搜索学生"
            onChange={(event) => setSearch(event.target.value)}
            placeholder="搜索姓名或学号"
            value={search}
          />
          <button className="secondary-button" type="submit">查询</button>
        </form>
        <button className="text-button" onClick={() => void load("", true)} type="button">查看含回收站记录</button>
      </section>

      {error && <p className="form-error" role="alert">{error}</p>}

      {showForm && (
        <section className="drawer-card">
          <div className="section-heading">
            <div><p className="eyebrow">STUDENT CARD</p><h3>{editing ? "编辑学生" : "新增学生"}</h3></div>
            <button className="icon-button" onClick={() => setShowForm(false)} type="button" aria-label="关闭">×</button>
          </div>
          <form className="student-form" onSubmit={submit}>
            <label>姓名<input name="name" defaultValue={editing?.name} required /></label>
            <label>性别
              <select name="gender" defaultValue={editing?.gender || ""}>
                <option value="">未填写</option><option>男</option><option>女</option>
              </select>
            </label>
            <label>学号<input name="studentNumber" defaultValue={editing?.studentNumber} /></label>
            <label className="wide">备注<textarea name="notes" defaultValue={editing?.notes} rows={2} /></label>
            <div className="wide form-actions">
              <button className="secondary-button" onClick={() => setShowForm(false)} type="button">取消</button>
              <button className="primary-button compact" type="submit">保存学生</button>
            </div>
          </form>
        </section>
      )}

      <section className="table-card">
        {page?.items.length ? (
          <table>
            <thead><tr><th>序号</th><th>姓名</th><th>性别</th><th>学号</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              {page.items.map((student, index) => (
                <tr className={student.deletedAt ? "trashed-row" : ""} key={student.id}>
                  <td>{String(index + 1).padStart(2, "0")}</td>
                  <td><strong>{student.name}</strong>{student.deletedAt && <small className="tag">回收站</small>}</td>
                  <td>{student.gender || "—"}</td>
                  <td>{student.studentNumber || "—"}</td>
                  <td>{student.status === "active" ? "在班" : student.status}</td>
                  <td className="row-actions">
                    {!student.deletedAt && <>
                      <button onClick={() => { setEditing(student); setShowForm(true); }} type="button">编辑</button>
                      <button className="danger-link" onClick={() => void trash(student)} type="button">删除</button>
                    </>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className="empty-table"><span>名</span><h3>还没有学生资料</h3><p>手动添加，或稍后从 Excel、JSON、CSV 导入。</p></div>
        )}
      </section>
    </>
  );
}
