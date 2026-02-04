export interface Event {
  type: string;
  actor?: string;
  message: string;
  data?: any;
}

export interface AgentStatus {
  status: string;
  isRemote: boolean;
}

export interface AgentStates {
    [name: string]: AgentStatus;
  }
  