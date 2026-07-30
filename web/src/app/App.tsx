import { useEffect, useState } from "react";

import { api, ApiClientError } from "../api/client";
import type { ClassInfo, StudentPage } from "../api/types";
import { AppShell, type Section } from "../components/AppShell";
import { ClassOverviewPage } from "../features/classes/ClassOverviewPage";
import { CreateClassPage } from "../features/classes/CreateClassPage";
import { CustomFieldsPage } from "../features/custom-fields/CustomFieldsPage";
import { DashboardPage } from "../features/dashboard/DashboardPage";
import { StudentListPage } from "../features/students/StudentListPage";

export function App() {
  const [currentClass, setCurrentClass] = useState<ClassInfo | null | undefined>(undefined);
  const [students, setStudents] = useState<StudentPage | null>(null);
  const [section, setSection] = useState<Section>("dashboard");
  const [openStudentForm, setOpenStudentForm] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    void loadWorkspace();
  }, []);

  async function loadWorkspace() {
    try {
      const current = await api.get<ClassInfo | null>("/classes/current");
      setCurrentClass(current);
      if (current) {
        const page = await api.get<StudentPage>(
          `/students?classId=${current.id}&page=1&pageSize=50`
        );
        setStudents(page);
      }
    } catch (reason) {
      setError(reason instanceof ApiClientError ? reason.message : "工作台加载失败");
      setCurrentClass(null);
    }
  }

  function navigate(next: Section) {
    setSection(next);
    setOpenStudentForm(false);
    window.history.replaceState({ section: next }, "", `#${next}`);
  }

  if (currentClass === undefined) {
    return (
      <main className="loading-screen">
        <div className="brand-seal large">班</div>
        <h1>ClassPilot 班主任工作台</h1>
        <p>正在打开本地班级底册…</p>
      </main>
    );
  }

  if (!currentClass) {
    return (
      <>
        {error && <p className="startup-error" role="alert">{error}</p>}
        <CreateClassPage onCreated={(created) => {
          setCurrentClass(created);
          setStudents({ items: [], total: 0, page: 1, pageSize: 50 });
        }} />
      </>
    );
  }

  return (
    <AppShell
      className={currentClass.name}
      onNavigate={navigate}
      section={section}
    >
      {section === "dashboard" && (
        <DashboardPage
          currentClass={currentClass}
          onAddStudent={() => {
            setOpenStudentForm(true);
            setSection("students");
          }}
          onViewStudents={() => navigate("students")}
          students={students}
        />
      )}
      {section === "class" && (
        <ClassOverviewPage onChanged={setCurrentClass} value={currentClass} />
      )}
      {section === "students" && (
        <StudentListPage
          currentClass={currentClass}
          initialPage={students}
          onChanged={setStudents}
          openCreateInitially={openStudentForm}
        />
      )}
      {section === "fields" && <CustomFieldsPage currentClass={currentClass} />}
      {section === "data" && (
        <ComingSoon
          eyebrow="DATA CENTER"
          title="数据中心"
          copy="下一批将接入 Excel、JSON、CSV 导入导出，以及完整备份与恢复。"
        />
      )}
      {section === "history" && (
        <ComingSoon
          eyebrow="ARCHIVE"
          title="历史班级"
          copy="当前班级归档后，会在这里以只读方式长期保存。"
        />
      )}
      {section === "settings" && (
        <ComingSoon
          eyebrow="LOCAL SETTINGS"
          title="工作台设置"
          copy="ClassPilot 当前仅在本机运行，数据保存在启动程序所在工作台目录。"
        />
      )}
    </AppShell>
  );
}

function ComingSoon({ eyebrow, title, copy }: { eyebrow: string; title: string; copy: string }) {
  return (
    <>
      <header className="page-header">
        <div><p className="eyebrow">{eyebrow}</p><h2>{title}</h2><p>{copy}</p></div>
      </header>
      <section className="paper-card empty-inline">
        <div className="brand-seal">班</div>
        <h3>基础入口已经预留</h3>
        <p>相关能力会在下一批实现后直接出现在这里。</p>
      </section>
    </>
  );
}
