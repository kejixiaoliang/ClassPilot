import { type FormEvent, useEffect, useState } from "react";

import { api, ApiClientError } from "../../api/client";
import type { ClassInfo, CustomField, FieldType } from "../../api/types";

export function CustomFieldsPage({ currentClass }: { currentClass: ClassInfo }) {
  const [items, setItems] = useState<CustomField[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [type, setType] = useState<FieldType>("text");
  const [error, setError] = useState("");

  useEffect(() => { void load(); }, [currentClass.id]);

  async function load() {
    try {
      const result = await api.get<{ items: CustomField[] }>(
        `/custom-fields?classId=${currentClass.id}&includeDisabled=true`
      );
      setItems(result.items);
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "加载字段失败");
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      await api.post("/custom-fields", {
        classId: currentClass.id,
        name: form.get("name"),
        type,
        description: form.get("description"),
        required: form.get("required") === "on",
        showInList: form.get("showInList") === "on",
        options: ["single_choice", "multi_choice"].includes(type)
          ? String(form.get("options") || "").split(/[，,\n]/).map((item) => item.trim()).filter(Boolean)
          : []
      });
      setShowForm(false);
      await load();
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "保存字段失败");
    }
  }

  async function disable(field: CustomField) {
    if (!window.confirm(`停用“${field.name}”？已有数据会继续保留。`)) return;
    await api.post(`/custom-fields/${field.id}/disable`);
    await load();
  }

  return (
    <>
      <header className="page-header">
        <div><p className="eyebrow">CUSTOM COLUMNS · 自定义信息</p><h2>字段设置</h2><p>不同学段需要的信息不一样，在这里扩展学生资料卡。</p></div>
        <button className="primary-button compact" onClick={() => setShowForm(true)} type="button">＋ 新建字段</button>
      </header>
      {error && <p className="form-error" role="alert">{error}</p>}
      {showForm && (
        <section className="drawer-card">
          <div className="section-heading"><h3>新建自定义字段</h3><button className="icon-button" onClick={() => setShowForm(false)} type="button">×</button></div>
          <form className="student-form" onSubmit={submit}>
            <label>字段名称<input name="name" placeholder="例如：住宿情况" required /></label>
            <label>字段类型
              <select value={type} onChange={(event) => setType(event.target.value as FieldType)}>
                <option value="text">短文本</option><option value="long_text">长文本</option>
                <option value="number">数字</option><option value="date">日期</option>
                <option value="single_choice">单选</option><option value="multi_choice">多选</option>
                <option value="boolean">是／否</option>
              </select>
            </label>
            <label className="wide">说明<input name="description" placeholder="告诉使用者应该填写什么" /></label>
            {["single_choice", "multi_choice"].includes(type) && (
              <label className="wide">选项<input name="options" placeholder="用逗号分隔，例如：走读，住校" required /></label>
            )}
            <label className="check-label"><input name="required" type="checkbox" /> 必填字段</label>
            <label className="check-label"><input name="showInList" type="checkbox" /> 在名单中显示</label>
            <div className="wide form-actions"><button className="primary-button compact" type="submit">保存字段</button></div>
          </form>
        </section>
      )}
      <section className="field-grid">
        {items.map((field) => (
          <article className={field.enabled ? "field-card" : "field-card disabled"} key={field.id}>
            <div><span className="field-type">{fieldTypeName(field.type)}</span>{field.required && <span className="required-mark">必填</span>}</div>
            <h3>{field.name}</h3><p>{field.description || "暂无说明"}</p>
            <footer><span>{field.showInList ? "名单中显示" : "仅详情显示"}</span>{field.enabled ? <button onClick={() => void disable(field)} type="button">停用</button> : <b>已停用 · 数据保留</b>}</footer>
          </article>
        ))}
        {!items.length && <div className="empty-table"><span>栏</span><h3>使用核心字段即可开始</h3><p>有额外需求时，再建立自定义字段。</p></div>}
      </section>
    </>
  );
}

function fieldTypeName(type: FieldType) {
  return {
    text: "短文本", long_text: "长文本", number: "数字", date: "日期",
    single_choice: "单选", multi_choice: "多选", boolean: "是／否"
  }[type];
}
