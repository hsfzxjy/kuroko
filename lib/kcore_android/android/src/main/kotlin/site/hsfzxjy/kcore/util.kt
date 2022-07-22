package site.hsfzxjy.kcore

import android.app.Activity
import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.embedding.engine.plugins.activity.ActivityAware
import io.flutter.plugin.common.MethodChannel
import kotlin.reflect.KProperty

class ResettableLazy<T>(private val initializer: () -> T) {
  private var _value: T? = null

  operator fun getValue(thisRef: Any?, p: KProperty<*>): T {
    if (_value == null) {
      _value = initializer()
    }
    return _value!!
  }

  operator fun setValue(thisRef: Any?, p: KProperty<*>, v: T) {
    _value = v
  }
}


interface FlutterPluginWithActivity : FlutterPlugin, ActivityAware {
  var activity: Activity?
}

class ResultSet {
  private val results = mutableSetOf<MethodChannel.Result>()

  fun add(result: MethodChannel.Result): Boolean {
    val isNew = results.isEmpty()
    results.add(result)
    return isNew
  }

  fun success(obj: Any?) {
    results.removeAll {
      it.success(obj)
      true
    }
  }

  fun error(errCode: String, errMsg: String) {
    results.removeAll {
      it.error(errCode, errMsg, null)
      true
    }
  }
}