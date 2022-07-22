package site.hsfzxjy.kcore

import android.annotation.SuppressLint
import android.app.Activity
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.util.Log
import androidx.core.app.ActivityCompat
import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.embedding.engine.plugins.activity.ActivityPluginBinding
import io.flutter.plugin.common.*


@SuppressLint("MissingPermission")
private class DiscoveryManager(val related: BluetoothRelated) {
  private var started = false
  private var sink: EventChannel.EventSink? = null
  private val receiver = object : BroadcastReceiver() {
    override fun onReceive(ctx: Context?, intent: Intent?) {
      when (intent!!.action!!) {
        BluetoothDevice.ACTION_FOUND -> {
          val device = intent.asBluetoothDevice
          Log.d(TAG, "found device ${device.address}")
          sink?.success(device.asMap())
        }
        BluetoothAdapter.ACTION_DISCOVERY_FINISHED ->
          stop()
      }
    }
  }
  private val channel: EventChannel by lazy {
    val c = EventChannel(
      related.messenger,
      "${BluetoothRelated.CLASS_NAMESPACE}/discovery",
    )
    c.setStreamHandler(object : EventChannel.StreamHandler {
      override fun onListen(arguments: Any?, events: EventChannel.EventSink?) {
        sink = events
      }

      override fun onCancel(arguments: Any?) = stop()
    })
    c
  }

  fun stop() {
    related.context!!.unregisterReceiverSilent(receiver)
    related.adapter!!.cancelDiscovery()

    sink?.endOfStream()
    sink = null
    started = false
  }

  fun start(): Boolean {
    if (started) throw LogicalError("discovery already started")
    started = true
    val filter = IntentFilter()
    channel // ensure channel initialized
    filter.addAction(BluetoothAdapter.ACTION_DISCOVERY_FINISHED)
    filter.addAction(BluetoothDevice.ACTION_FOUND)
    related.context!!.registerReceiver(receiver, filter)
    related.adapter!!.startDiscovery()
    return true
  }

  companion object {
    private const val TAG = "DiscoveryManager"
  }
}

@SuppressLint("MissingPermission")
private class BondManager(val related: BluetoothRelated) {
  private var started = false
  private var device: BluetoothDevice? = null
  private var result: MethodChannel.Result? = null
  private var receiver = object : BroadcastReceiver() {
    override fun onReceive(ctx: Context?, intent: Intent?) {
      if (!started || intent!!.action != BluetoothDevice.ACTION_BOND_STATE_CHANGED) return

      val dev = intent.asBluetoothDevice
      if (dev != device) return

      val bondState = intent.getIntExtra(BluetoothDevice.EXTRA_BOND_STATE, BluetoothDevice.ERROR)
      when (bondState) {
        BluetoothDevice.BOND_BONDING -> return
        BluetoothDevice.BOND_BONDED -> result!!.success(true)
        BluetoothDevice.BOND_NONE -> result!!.success(false)
        else -> {
          result!!.raises(LogicalError("unknown bond state $bondState"))
          return
        }
      }
      cancel()
    }
  }


  private fun cancel() {
    related.context!!.unregisterReceiver(receiver)
    result = null
    device = null
    started = false
  }

  fun start(result: MethodChannel.Result, dev: BluetoothDevice) {
    if (started) throw LogicalError("bond process already started")
    started = true
    device = dev
    this.result = result
    when (dev.bondState) {
      BluetoothDevice.BOND_BONDED, BluetoothDevice.BOND_BONDING ->
        throw LogicalError("device bonding/already bonded")
    }
    related.context!!.registerReceiver(receiver,
      IntentFilter(BluetoothDevice.ACTION_BOND_STATE_CHANGED))
    if (!dev.createBond()) {
      cancel()
      result.success(false)
    } else result.success(true)
  }

  fun stop() {
    device?.removeBond()
    val res = result
    cancel()
    res?.raises(CanceledError())
  }
}

const val REQUEST_ENABLE_BLUETOOTH: Int = 0

@SuppressLint("MissingPermission")
class BluetoothRelated : FlutterPluginWithActivity, MethodChannel.MethodCallHandler {
  lateinit var messenger: BinaryMessenger
  override var activity: Activity? = null
    set(value) {
      context = null
      adapter = null
      field = value
    }
  var context: Context? by ResettableLazy {
    activity!!.applicationContext
  }
  var adapter: BluetoothAdapter? by ResettableLazy {
    context!!.bluetoothAdapter
  }
  private lateinit var channel: MethodChannel

  private var discovery = DiscoveryManager(this)
  private var bondManager = BondManager(this)

  private val pendingResults = mapOf<Int, ResultSet>().withDefault { ResultSet() }

  override fun onAttachedToEngine(binding: FlutterPlugin.FlutterPluginBinding) {
    messenger = binding.binaryMessenger
    channel = MethodChannel(messenger, CLASS_NAMESPACE)
    channel.setMethodCallHandler(this)
  }

  override fun onAttachedToActivity(binding: ActivityPluginBinding) {
    binding.addActivityResultListener(object : PluginRegistry.ActivityResultListener {
      override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?): Boolean {
        when (requestCode) {
          REQUEST_ENABLE_BLUETOOTH -> {
            pendingResults[requestCode]?.success(resultCode != 0)
            return true
          }
        }
        return true
      }
    })
  }

  override fun onMethodCall(call: MethodCall, result: MethodChannel.Result) =
    result.collect {
      when (call.method) {
        "ensureEnabled" -> {
          if (adapter!!.isEnabled) {
            result.success(true)
            return@collect
          }

          if (pendingResults[REQUEST_ENABLE_BLUETOOTH]!!.add(result)) {
            val intent = Intent(BluetoothAdapter.ACTION_REQUEST_ENABLE)
            ActivityCompat.startActivityForResult(activity!!,
              intent,
              REQUEST_ENABLE_BLUETOOTH,
              null)
          }
        }
        "startDiscovery" -> {
          Log.d(TAG, "startDiscovery")
          discovery.start()
          result.success(null)
        }
        "stopDiscovery" -> {
          discovery.stop()
          result.success(null)
        }
        "getDevice" -> {
          val addr = call.argBluetoothAddress("address", true)!!
          result.success(adapter!!.getRemoteDevice(addr).asMap())
        }
        "getBondState" -> {
          val addr = call.argBluetoothAddress("address", true)!!
          result.success(adapter!!.getRemoteDevice(addr).bondState)
        }
        "startBond" -> {
          val addr = call.argBluetoothAddress("address", true)!!
          bondManager.start(result, adapter!!.getRemoteDevice(addr))
        }
        "stopBond" -> {
          bondManager.stop()
          result.success(null)
        }
        else -> result.notImplemented()
      }
    }


  override fun onDetachedFromEngine(binding: FlutterPlugin.FlutterPluginBinding) {
    channel.setMethodCallHandler(null)
  }

  override fun onDetachedFromActivity() {}
  override fun onDetachedFromActivityForConfigChanges() {}
  override fun onReattachedToActivityForConfigChanges(binding: ActivityPluginBinding) {}

  companion object {
    const val CLASS_NAMESPACE = "$PLUGIN_NAMESPACE/bluetooth"
    private const val TAG = "BluetoothRelated"
  }
}