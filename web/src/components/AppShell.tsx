import type { ReactNode } from "react";

export type Section =
  | "dashboard"
  | "class"
  | "students"
  | "fields"
  | "data"
  | "history"
  | "settings";

type Props = {
  className: string;
  section: Section;
  onNavigate: (section: Section) => void;
  children: ReactNode;
};

const groups: Array<{ label: string; items: Array<{ id: Section; label: string; mark: string }> }> = [
  {
    label: "今日班务",
    items: [
      { id: "dashboard", label: "工作台首页", mark: "今" },
      { id: "students", label: "学生名单", mark: "生" }
    ]
  },
  {
    label: "班级底册",
    items: [
      { id: "class", label: "班级概览", mark: "班" },
      { id: "fields", label: "字段设置", mark: "栏" }
    ]
  },
  {
    label: "资料管理",
    items: [
      { id: "data", label: "数据中心", mark: "数" },
      { id: "history", label: "历史班级", mark: "档" },
      { id: "settings", label: "设置", mark: "设" }
    ]
  }
];

export function AppShell({ className, section, onNavigate, children }: Props) {
  return (
    <div className="app-frame">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-seal">班</div>
          <div>
            <p className="brand-en">CLASSPILOT</p>
            <h1>班主任工作台</h1>
          </div>
        </div>

        <div className="current-class-chip">
          <span className="status-dot" />
          <div>
            <small>当前班级</small>
            <strong>{className}</strong>
          </div>
        </div>

        <nav aria-label="主导航">
          {groups.map((group) => (
            <div className="nav-group" key={group.label}>
              <p>{group.label}</p>
              {group.items.map((item) => (
                <button
                  className={section === item.id ? "nav-item active" : "nav-item"}
                  key={item.id}
                  onClick={() => onNavigate(item.id)}
                  type="button"
                >
                  <span aria-hidden="true">{item.mark}</span>
                  {item.label}
                </button>
              ))}
            </div>
          ))}
        </nav>

        <footer className="sidebar-footer">
          <span>本地离线</span>
          <span>数据仅在此电脑</span>
        </footer>
      </aside>
      <main className="content">{children}</main>
    </div>
  );
}
