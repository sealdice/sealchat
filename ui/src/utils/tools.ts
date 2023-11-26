import { memoize } from "lodash-es";

export function formDataToJson(formData: FormData): Record<string, any> {
  const jsonObject: Record<string, any> = {};

  formData.forEach((value, key) => {
    const keys = key.split('.');
    let currentObject = jsonObject;

    for (let i = 0; i < keys.length - 1; i++) {
      const currentKey = keys[i];
      currentObject[currentKey] = currentObject[currentKey] || {};
      currentObject = currentObject[currentKey];
    }

    currentObject[keys[keys.length - 1]] = value.toString();
  });

  return jsonObject;
}

export function blobToArrayBuffer(blob: Blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      resolve(reader.result);
    };
    reader.onerror = reject;
    reader.readAsArrayBuffer(blob);
  }) as Promise<string | ArrayBuffer | null>;
}

export function dataURItoBlob(dataURI: string) {
  // convert base64 to raw binary data held in a string
  // doesn't handle URLEncoded DataURIs - see SO answer #6850276 for code that does this
  var byteString = atob(dataURI.split(',')[1])

  // separate out the mime component
  var mimeString = dataURI.split(',')[0].split(':')[1].split(';')[0]

  // write the bytes of the string to an ArrayBuffer
  var ab = new ArrayBuffer(byteString.length)
  var ia = new Uint8Array(ab)
  for (var i = 0; i < byteString.length; i++) {
    ia[i] = byteString.charCodeAt(i)
  }

  return new Blob([ab], { type: mimeString })
}

export function memoizeWithTimeout<T>(func: (...args: any[]) => T, timeout: number): (...args: any[]) => T {
  const memoizedFunc = memoize(func);

  function timedMemoizedFunc(...args: any[]): T {
    const result = memoizedFunc(...args);
    setTimeout(() => {
      memoizedFunc.cache.delete(args.toString());
    }, timeout);
    return result;
  }

  return timedMemoizedFunc;
}
