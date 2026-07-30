export type ClassInfo = {
  id: string;
  name: string;
  stage: string;
  grade: string;
  schoolYear: string;
  status: "current" | "archived";
  notes: string;
  createdAt: string;
  updatedAt: string;
  archivedAt?: string;
};

export type Student = {
  id: string;
  classId: string;
  name: string;
  gender: string;
  studentNumber: string;
  birthDate?: string;
  enrollmentDate?: string;
  status: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
};

export type StudentPage = {
  items: Student[];
  total: number;
  page: number;
  pageSize: number;
};

export type FieldType =
  | "text"
  | "long_text"
  | "number"
  | "date"
  | "single_choice"
  | "multi_choice"
  | "boolean";

export type CustomField = {
  id: string;
  classId: string;
  name: string;
  type: FieldType;
  description: string;
  required: boolean;
  enabled: boolean;
  showInList: boolean;
  options: string[];
  sortOrder: number;
};
