import Element from '@satorijs/element'
import { urlBase } from '@/stores/_config';
import DOMPurify from 'dompurify';

// 备选方案，移动解析部分到这里
// export const parseContent = (props: any, ) => {
//   const content = props.content;
//   const items = Element.parse(content);
//   let textItems = []

//   for (const item of items) {
//     switch (item.type) {
//       case 'img':
//         if (item.attrs.src && item.attrs.src.startsWith('id:')) {
//           item.attrs.src = item.attrs.src.replace('id:', `${urlBase}/api/v1/attachments/`);
//         }
//         textItems.push(DOMPurify.sanitize(item.toString()));
//         hasImage.value = true;
//         break;
//       case "at":
//         if (item.attrs.id == user.info.id) {
//           textItems.push(`<span class="text-blue-500 bg-gray-400 px-1" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
//         } else {
//           textItems.push(`<span class="text-blue-500" style="white-space: pre-wrap">@${item.attrs.name}</span>`);
//         }
//       default:
//         textItems.push(`<span style="white-space: pre-wrap">${item.toString()}</span>`);
//         break;
//     }
//   }

//   return textItems.join('');
// }
