import { type FormEvent, useState } from "react";

import { api, ApiClientError } from "../../api/client";
import type { ClassInfo } from "../../api/types";

type Props = {
  onCreated: (value: ClassInfo) => void;
};

export function CreateClassPage({ onCreated }: Props) {
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const created = await api.post<ClassInfo>("/classes/current", {
        name: form.get("name"),
        stage: form.get("stage"),
        grade: form.get("grade"),
        schoolYear: form.get("schoolYear"),
        notes: form.get("notes")
      });
      onCreated(created);
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "创建班级失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="onboarding">
      <h1 className="sr-only">ClassPilot 班主任工作台</h1>
      <section className="onboarding-copy">
        <div className="brand-seal large">班</div>
        <p className="eyebrow">CLASSPILOT · LOCAL FIRST</p>
        <h1>把一个班级，<br />安稳地放在电脑里。</h1>
        <p>
          无需登录，也无需联网。先建立当前班级，之后所有学生资料、附件和备份都会留在这套工作台目录中。
        </p>
        <div className="promise-list">
          <span>本地 SQLite</span>
          <span>可整目录迁移</span>
          <span>自动备份提醒</span>
        </div>
      </section>

      <section className="onboarding-card">
        <p className="step-mark">01 / 建立底册</p>
        <h2>创建第一个班级</h2>
        <p className="muted">这些信息以后都可以修改。</p>
        <form onSubmit={submit}>
          <label>
            班级名称
            <input name="name" placeholder="例如：七年级一班" required />
          </label>
          <div className="form-grid">
            <label>
              学段
              <select name="stage" defaultValue="初中">
                <option>小学</option>
                <option>初中</option>
                <option>高中</option>
                <option>中职／高职</option>
                <option>其他</option>
              </select>
            </label>
            <label>
              年级
              <input name="grade" placeholder="七年级" required />
            </label>
          </div>
          <label>
            学年
            <input name="schoolYear" defaultValue="2026-2027" required />
          </label>
          <label>
            备注
            <textarea name="notes" placeholder="可以先留空" rows={3} />
          </label>
          {error && <p className="form-error" role="alert">{error}</p>}
          <button className="primary-button" disabled={saving} type="submit">
            {saving ? "正在创建…" : "建立班级底册"}
          </button>
        </form>
      </section>
    </main>
  );
}
