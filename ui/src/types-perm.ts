import type { ChannelRolePermSheet } from "./types-perm-channel";
import type { SystemRolePermSheet } from "./types-perm-system";

export enum PermResult {
  UNSET = 0,
  ALLOWED = 1,
  DENIED = 2
}

export type PermCheckKey = keyof SystemRolePermSheet | keyof ChannelRolePermSheet;

export interface PermTreeNode {
  name: string;
  modelName?: PermCheckKey,
  children?: PermTreeNode[];
}
