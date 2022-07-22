package site.hsfzxjy.kcore

import android.app.Activity
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothManager
import android.content.Context
import android.util.Log
import androidx.annotation.NonNull

import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.embedding.engine.plugins.activity.ActivityAware
import io.flutter.embedding.engine.plugins.activity.ActivityPluginBinding
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import io.flutter.plugin.common.MethodChannel.MethodCallHandler
import io.flutter.plugin.common.MethodChannel.Result

const val PLUGIN_NAMESPACE = "kcore"


class KcoreAndroidPlugin : FlutterPlugin, MethodCallHandler, ActivityAware {
  private lateinit var channel: MethodChannel
  private lateinit var activity: Activity
  private lateinit var appContext: Context
  private val childPlugins: List<FlutterPluginWithActivity> = listOf(
    BluetoothRelated()
  )

  override fun onMethodCall(@NonNull call: MethodCall, @NonNull result: Result) =
    when (call.method) {
      "initDLL" -> {
        val ret = initDLL(activity, appContext.bluetoothManager, appContext.bluetoothAdapter)
        if (ret != 0)
          result.error("init_error", "initDLL() returns $ret", null)
        else
          result.success(null)
      }
      else ->
        result.notImplemented()
    }

  override fun onAttachedToEngine(@NonNull flutterPluginBinding: FlutterPlugin.FlutterPluginBinding) {
    Log.d(TAG, "attached to engine")
    channel = MethodChannel(flutterPluginBinding.binaryMessenger, PLUGIN_NAMESPACE)
    channel.setMethodCallHandler(this)
    appContext = flutterPluginBinding.applicationContext

    childPlugins.forEach { it.onAttachedToEngine(flutterPluginBinding) }
//    throw Error()
  }

  override fun onDetachedFromEngine(@NonNull binding: FlutterPlugin.FlutterPluginBinding) {
    channel.setMethodCallHandler(null)
    childPlugins.forEach { it.onDetachedFromEngine(binding) }
  }

  override fun onAttachedToActivity(binding: ActivityPluginBinding) {
    Log.d(TAG, "attached to activity")
    this.activity = binding.activity
    childPlugins.forEach {
      it.activity = activity
      it.onAttachedToActivity(binding)
    }
  }

  override fun onDetachedFromActivity() {}
  override fun onDetachedFromActivityForConfigChanges() {}
  override fun onReattachedToActivityForConfigChanges(binding: ActivityPluginBinding) {}

  private external fun initDLL(
    act: Activity,
    btMan: BluetoothManager,
    btAdapter: BluetoothAdapter,
  ): Int

  companion object {
    private const val TAG = "KcoreAndroidPlugin"

    init {
      System.loadLibrary("kcore")
    }
  }
}
