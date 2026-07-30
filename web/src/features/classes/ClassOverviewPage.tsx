import { type FormEvent, useState } from "react";

import { api, ApiClientError } from "../../api/client";
import type { ClassInfo } from "../../api/types";

export function ClassOverviewPage({ value, onChanged }: { value: ClassInfo; onChanged: (value: ClassInfo) => void }) {
  const [error, setError] = useState("");
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const updated = await api.patch<ClassInfo>("/classes/current", {
        name: form.get("name"), stage: form.get("stage"), grade: form.get("grade"),
        schoolYear: form.get("schoolYear"), notes: form.get("notes")
      });
      onChanged(updated);
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "保存班级失败");
    }
  }
  return (
    <>
      <header className="page-header"><div><p className="eyebrow">CLASS PROFILE · 班级底册</p><h2>班级概览</h2><p>维护当前班级的基础信息。</p></div></header>
      <section className="paper-card form-card">
        <form onSubmit={submit}>
          <div className="form-grid">
            <label>班级名称<input name="name" defaultValue={value.name} required /></label>
            <label>学段<input name="stage" defaultValue={value.stage} required /></label>
            <label>年级<input name="grade" defaultValue={value.grade} required /></label>
            <label>学年<input name="schoolYear" defaultValue={value.schoolYear} required /></label>
          </div>
          <label>备注<textarea name="notes" defaultValue={value.notes} rows={5} /></label>
          {error && <p className="form-error" role="alert">{error}</p>}
          <button className="primary-button compact" type="submit">保存班级资料</button>
        </form>
      </section>
    </>
  );
}
