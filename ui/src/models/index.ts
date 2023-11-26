import Dexie, { type Table } from 'dexie';
import { urlBase } from '@/stores/_config';

export interface Thumb {
  id?: string;
  filename: string;
  recentUsed: number;
  data: string | ArrayBuffer | null;
  mimeType: string;
}

export class MySubClassedDexie extends Dexie {
  // 'friends' is added by dexie when declaring the stores()
  // We just tell the typing system this is the case
  thumbs!: Table<Thumb>;

  constructor() {
    super('myDatabase');
    this.version(1).stores({
      thumbs: '++id, recentUsed, filename, data, mimeType' // Primary key and indexed props
    });
  }
}


export function getSrc(i: Thumb) {
  if (i.data) {
    let URL = window.URL || window.webkitURL
    if (URL && URL.createObjectURL) {
      const b = new Blob([i.data as any], { type: i.mimeType })
      return URL.createObjectURL(b)
    }
  } else {
      return `${urlBase}/api/v1/attachments/${i.id}`;
  }
}

export const db = new MySubClassedDexie();
