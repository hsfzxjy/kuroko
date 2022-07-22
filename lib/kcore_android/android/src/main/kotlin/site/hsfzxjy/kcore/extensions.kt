package site.hsfzxjy.kcore

import android.annotation.SuppressLint
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothManager
import android.content.BroadcastReceiver
import android.content.Intent
import android.content.Context
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import java.lang.ClassCastException
import java.lang.Exception
import java.lang.IllegalArgumentException
import java.util.*

@SuppressLint("MissingPermission")
fun BluetoothDevice.asMap(): Map<String, Any> =
  mapOf(
    "address" to this.address,
    "name" to this.name,
    "type" to this.type,
    "bondState" to this.bondState,
  )

fun BluetoothDevice.removeBond(): Boolean {
  val method = this.javaClass.getMethod("removeBond")
  return method.invoke(this) as Boolean
}

val Context.bluetoothAdapter: BluetoothAdapter
  get() = this.bluetoothManager.adapter

val Context.bluetoothManager: BluetoothManager
  get() = this.getSystemService(Context.BLUETOOTH_SERVICE) as BluetoothManager

val Intent.asBluetoothDevice: BluetoothDevice
  get() = getParcelableExtra(BluetoothDevice.EXTRA_DEVICE)!!

fun Context.unregisterReceiverSilent(receiver: BroadcastReceiver) {
  try {
    unregisterReceiver(receiver)
  } catch (_: IllegalArgumentException) {
  }
}

internal abstract class MethodCallError(val msg: String) : Exception() {
  abstract val errorCode: String
}

internal class ArgError(msg: String) : MethodCallError(msg) {
  override val errorCode = "arg_error"

  companion object {
    fun invalidType(key: String): Nothing = throw ArgError("argument '$key' has incorrect type")
  }
}

internal class LogicalError(msg: String) : MethodCallError(msg) {
  override val errorCode = "logical_error"
}

internal class CanceledError : MethodCallError("method call has been canceled") {
  override val errorCode = "canceled_error"
}

internal fun MethodChannel.Result.raises(ex: MethodCallError) {
  this.error(ex.errorCode, ex.msg, ex)
}

fun MethodChannel.Result.collect(block: () -> Unit) {
  try {
    block()
  } catch (ex: MethodCallError) {
    this.error(ex.errorCode, ex.msg, ex)
  }
}

fun <T> MethodCall.arg(key: String, required: Boolean): T? {
  if (required && !this.hasArgument(key)) throw ArgError("argument '$key' is required")
  try {
    return this.argument<T?>(key)
  } catch (ex: ClassCastException) {
    ArgError.invalidType(key)
  }
}

fun <T> MethodCall.argRequired(key: String): T =
  Objects.requireNonNull<T>(arg<T>(key, true))

fun MethodCall.argUUID(key: String, required: Boolean): UUID? {
  try {
    return UUID.fromString(arg<String>(key, required) ?: return null)
  } catch (ex: IllegalArgumentException) {
    ArgError.invalidType(key)
  }
}

fun MethodCall.argBluetoothAddress(key: String, required: Boolean): String? {
  val ret = arg<String>(key, required) ?: return null
  if (!BluetoothAdapter.checkBluetoothAddress(ret)) ArgError.invalidType(key)
  return ret
}