import type { ClassInfo, StudentPage } from "../../api/types";

type Props = {
  currentClass: ClassInfo;
  students: StudentPage | null;
  onAddStudent: () => void;
  onViewStudents: () => void;
};

export function DashboardPage({ currentClass, students, onAddStudent, onViewStudents }: Props) {
  const total = students?.total ?? 0;
  const boys = students?.items.filter((item) => item.gender === "男").length ?? 0;
  const girls = students?.items.filter((item) => item.gender === "女").length ?? 0;

  return (
    <>
      <header className="page-header">
        <div>
          <p className="eyebrow">CURRENT CLASS · 当前班级</p>
          <h2>{currentClass.name}</h2>
          <p>{currentClass.schoolYear} 学年 · {currentClass.stage} · {currentClass.grade}</p>
        </div>
        <button className="primary-button compact" onClick={onAddStudent} type="button">＋ 添加学生</button>
      </header>

      <section className="hero-ledger">
        <div>
          <p>班级人数</p>
          <strong>{total}</strong>
          <span>人</span>
        </div>
        <div className="ledger-stat">
          <span>男生</span><b>{boys}</b>
          <span>女生</span><b>{girls}</b>
          <span>未填写</span><b>{Math.max(total - boys - girls, 0)}</b>
        </div>
        <button className="text-button" onClick={onViewStudents} type="button">查看完整名单 →</button>
      </section>

      <div className="dashboard-grid">
        <section className="paper-card">
          <div className="section-heading">
            <div><p className="eyebrow">RECENT</p><h3>最近学生</h3></div>
            <button className="text-button" onClick={onViewStudents} type="button">全部</button>
          </div>
          {students?.items.length ? (
            <ul className="recent-list">
              {students.items.slice(0, 6).map((student, index) => (
                <li key={student.id}>
                  <span className="number-mark">{String(index + 1).padStart(2, "0")}</span>
                  <strong>{student.name}</strong>
                  <small>{student.studentNumber || "未填写学号"}</small>
                </li>
              ))}
            </ul>
          ) : (
            <div className="empty-inline">
              <p>花名册还是空的。</p>
              <button className="text-button" onClick={onAddStudent} type="button">添加第一位学生</button>
            </div>
          )}
        </section>

        <section className="paper-card reminder-card">
          <p className="eyebrow">DATA SAFETY</p>
          <h3>数据安全</h3>
          <div className="backup-ring"><span>尚未</span><strong>备份</strong></div>
          <p>第一次录入名单后，建议立即创建完整备份。</p>
          <button className="secondary-button" type="button" disabled>备份功能将在下一批完成</button>
        </section>
      </div>
    </>
  );
}
