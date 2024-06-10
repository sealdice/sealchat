import type { User, Message, Guild, GuildMember, Opcode, GatewayPayloadStructure, Channel } from '@satorijs/protocol'

export interface SatoriMessage {
  id?: string;
  channel?: Channel;
  guild?: Guild;
  user?: User;
  member?: GuildMember;
  content?: string;
  elements?: any[]; // Element[] 这个好像会让vscode提示一个错误
  timestamp?: number;
  quote?: SatoriMessage;
  createdAt?: number;
  updatedAt?: number;
}

export interface ServerConfig {
  serveAt: string;
  domain: string;
  registerOpen: boolean;
  webUrl: string;
  chatHistoryPersistentDays: number;
  imageSizeLimit: number;
  imageCompress: boolean;
}

export interface UserInfo {
  id: string;
  createdAt: null | string;
  updatedAt: null | string;
  deletedAt: null | string;
  username: string;
  nick: string;
  avatar: string;
  brief: string;
  role: "role-admin" | 'role-disabled' | '';
}

export interface TalkMessage {
  id: string;
  time: number;
  name: string;
  content: string;
  isMe?: boolean;
  raw?: any;
}

// https://satori.js.org/zh-CN/resources/message.html#%E5%8F%91%E9%80%81%E6%B6%88%E6%81%AF
interface APIMessageCreate {
  // api: 'message.create'
  channel_id: string
  content: string
}

export type APIMessageCreateResp = Message

interface APIMessageGet {
  api: 'message.get'
  channel_id: string
  message_id: string
}

// 扩展部分
interface APIChannelCreate {
  api: 'channel.create'
  name: string
}

interface APIChannelList {
  // api: 'channel.list'
}


export interface APIChannelCreateResp {
  id: string
  name: string
  parent_id: string
  // type
}

export interface APIChannelListResp {
  data: Channel[],
  echo?: string,
  next?: string,
}

export type APIMessage = APIMessageCreate | APIMessageGet | APIChannelList;

export interface UserEmojiModel {
  id: string
  attachmentId: string;
  order?: number;
}
