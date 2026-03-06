export type Action =
  | 'create'
  | 'update'
  | 'delete'
  | 'replace'
  | 'create-before-delete'
  | 'delete-before-create';

export type RiskLevel = 'high' | 'medium' | 'low';

export interface AttributeChange {
  name: string;
  old_value: unknown;
  new_value: unknown;
  computed?: boolean;
}

export interface ResourceChange {
  address: string;
  type: string;
  name: string;
  provider_name: string;
  action: Action;
  action_reason?: string;
  attributes?: AttributeChange[];
  risk_level: RiskLevel;
  risk_reasons?: string[];
}

export interface PlanSummary {
  total_changes: number;
  adds: number;
  changes: number;
  destroys: number;
  replaces: number;
  high_risk: number;
  medium_risk: number;
  low_risk: number;
}

export interface Plan {
  format_version: string;
  terraform_version: string;
  resource_changes: ResourceChange[];
  summary: PlanSummary;
  plan_file?: string;
  working_dir?: string;
  timestamp: string;
}

export interface ApplyLine {
  text: string;
  timestamp: number;
  type: 'stdout' | 'stderr' | 'status';
}

export interface AIAnalysisData {
  findings: string[];
  risk_summary: string;
  recommendations: string[];
}
