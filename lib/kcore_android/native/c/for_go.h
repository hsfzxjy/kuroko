#ifndef _FOR_GO_H_
#define _FOR_GO_H_
#include "./dffi.h"

typedef enum {
  DartValue_kBool = 0,
  DartValue_kInt32,
  DartValue_kInt64,
  DartValue_kUint8Array,
  DartValue_kString,
  DartValue_kNull,
} DartValueKind;

typedef struct {
  intptr_t length;
  uint8_t* values;
} DartValue_vUint8Array;

typedef struct {
  DartValueKind kind;
  union {
    int32_t as_null;
    int32_t as_bool;
    int32_t as_int32;
    int64_t as_int64;
    DartValue_vUint8Array as_uint8_array;
    char* as_string;
  } value;
} DartValue;

extern void QueueCallback(DartCallback cb, int narg, DartValue* args);
#endif