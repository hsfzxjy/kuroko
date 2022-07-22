#include <android/log.h>
#include <android/looper.h>
#include <errno.h>
#include <jni.h>
#include <stdint.h>
#include <stdlib.h>
#include <unistd.h>

#include "./for_go.h"

#define LOGI(...) __android_log_print(ANDROID_LOG_INFO, "KCore_C", __VA_ARGS__)

void FreeObject(void* _, void* ptr) { free(ptr); }

#include "dart_api.h"
#include "dart_api_dl.h"
#include "dart_native_api.h"

static Dart_Port sendPort;

DART_EXPORT void CFree(void* ptr) { free(ptr); }

DART_EXPORT intptr_t InitializeDartFFI(void* data, Dart_Port port) {
  LOGI("InitializeDartFFI()");
  sendPort = port;
  return Dart_InitializeApiDL(data);
}

void DartValueToCObject(DartValue* arg, Dart_CObject* obj) {
  switch (arg->kind) {
    case DartValue_kNull:
      obj->type = Dart_CObject_kNull;
      break;
    case DartValue_kBool:
      obj->type = Dart_CObject_kBool;
      obj->value.as_bool = arg->value.as_bool;
      break;
    case DartValue_kInt32:
      obj->type = Dart_CObject_kInt32;
      obj->value.as_int32 = arg->value.as_int32;
      break;
    case DartValue_kInt64:
      obj->type = Dart_CObject_kInt64;
      obj->value.as_int64 = arg->value.as_int64;
      break;
    case DartValue_kString:
      obj->type = Dart_CObject_kString;
      obj->value.as_string = arg->value.as_string;
      break;
    case DartValue_kUint8Array:
      obj->type = Dart_CObject_kExternalTypedData;
      obj->value.as_external_typed_data.type = Dart_TypedData_kUint8;
      obj->value.as_external_typed_data.length =
          arg->value.as_uint8_array.length;
      obj->value.as_external_typed_data.data =
          obj->value.as_external_typed_data.peer =
              arg->value.as_uint8_array.values;
      obj->value.as_external_typed_data.callback = &FreeObject;
      break;
  }
}

const size_t MAX_NARGS = 15;

void QueueCallback(DartCallback cb, int narg, DartValue* args) {
  assert(narg <= MAX_NARGS);
  Dart_CObject* objptrs[MAX_NARGS + 1];
  Dart_CObject objs[MAX_NARGS + 1];

  objs[0].type = Dart_CObject_kInt64;
  objs[0].value.as_int64 = (int64_t)(cb);
  objptrs[0] = &objs[0];

  int i;
  for (i = 0; i < narg; i++) {
    DartValueToCObject(&args[i], &objs[i + 1]);
    objptrs[i + 1] = &objs[i + 1];
  }

  Dart_CObject obj;
  obj.type = Dart_CObject_kArray;
  obj.value.as_array.length = narg + 1;
  obj.value.as_array.values = &objptrs[0];

  Dart_PostCObject_DL(sendPort, &obj);
}

#include "dart_api_dl.c"
