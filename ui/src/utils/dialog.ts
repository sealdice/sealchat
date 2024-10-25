import { useDialog } from "naive-ui";
import type { DialogApiInjection } from "naive-ui/es/dialog/src/DialogProvider";


export async function dialogAskConfirm(dialog: DialogApiInjection, title = '是否进行此操作？', content = '请确认操作无误') {
  // const dialog = useDialog();
  return new Promise((resolve) => {
    dialog.warning({
      title: title,
      content: content,
      positiveText: '是的',
      negativeText: '取消',
      onPositiveClick: () => {
        resolve(true);
      },
      onNegativeClick: () => {
        resolve(false);
      }
    });
  });
}
