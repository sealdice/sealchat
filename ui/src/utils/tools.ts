
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
