part of kcore_android;

abstract class EnumLike<T> extends Enum {
  T get value;
}

Map<T, U> _enumLookupMap<T, U extends EnumLike<T>>(List<U> values) =>
    //ignore:prefer_for_elements_to_map_fromiterable
    Map.fromIterable(
      values,
      key: (x) => x.value,
      value: (x) => x,
    );

extension MethodChannelExt on MethodChannel {
  Future<T> invokeMethodNonNull<T>(String method, [dynamic arguments]) async {
    final result = await invokeMethod(method, arguments);
    return result!;
  }
}
