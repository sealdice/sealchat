import { AxiosError } from "axios";

export async function coverErrorMessage(inside: Function, dialogError: any, statusHook?: (status: number) => boolean | undefined) {
  try {
    return await inside();
  } catch (e) {
    if (e instanceof AxiosError) {
      let msg = e.response?.data?.message;
      if (statusHook) {
        if (statusHook(e.response?.status ?? 0)) {
          // 如果返回true，代表solved，不走其他流程
          return;
        }
      }
      switch (e.response?.status) {
        case 400: // StatusBadRequest
          dialogError('提示', `无法进行此操作` + (msg ? `: ${msg}` : ''));
          return;
        case 401: // StatusUnauthorized
          dialogError('提示', `权限错误` + (msg ? `: ${msg}` : ''), 'info');
          return;
        case 404: // StatusNotFound
          dialogError('提示', `资源不存在` + (msg ? `: ${msg}` : ''));
          return;
        default:
          dialogError('提示', `请求超时，未知问题` + (msg ? `: ${msg}` : ''));
          return;
      }
    }
    throw e;
  }
}
