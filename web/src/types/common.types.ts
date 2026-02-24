export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface SelectOption {
  label: string;
  value: string;
}

export interface FilterParams {
  field: string;
  operator: "eq" | "ne" | "gt" | "gte" | "lt" | "lte" | "contains" | "in";
  value: string | number | boolean | string[];
}

export interface SortParams {
  field: string;
  order: "asc" | "desc";
}

export interface TableParams {
  pagination: {
    current: number;
    pageSize: number;
  };
  filters?: FilterParams[];
  sorters?: SortParams[];
}

export interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ReactNode;
}

export interface KPIData {
  title: string;
  value: number | string;
  prefix?: string;
  suffix?: string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  icon?: React.ReactNode;
  color?: string;
}

export interface ChartDataPoint {
  label: string;
  value: number;
  color?: string;
}

export type StatusType =
  | "success"
  | "warning"
  | "error"
  | "info"
  | "default"
  | "processing";
