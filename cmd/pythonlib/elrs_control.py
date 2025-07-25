import ctypes
import os
import sys
from typing import Optional, Callable, List

# Load the shared library
lib_name = "elrs_control.so"
lib_path = os.path.join(os.path.dirname(__file__), lib_name)
lib = ctypes.CDLL(lib_path)

# Define callback types
LinkStatsCallback = ctypes.CFUNCTYPE(None, ctypes.c_int32, ctypes.c_int32, ctypes.c_uint32, ctypes.c_int32)
BatteryCallback = ctypes.CFUNCTYPE(None, ctypes.c_float, ctypes.c_float, ctypes.c_float)
GPSCallback = ctypes.CFUNCTYPE(None, ctypes.c_float, ctypes.c_float, ctypes.c_int32, ctypes.c_uint32, ctypes.c_float)
AttitudeCallback = ctypes.CFUNCTYPE(None, ctypes.c_float, ctypes.c_float, ctypes.c_float)

# Configure function signatures
lib.elrs_init.argtypes = [ctypes.c_char_p, ctypes.c_int]
lib.elrs_init.restype = ctypes.c_int

lib.elrs_close.argtypes = []
lib.elrs_close.restype = None

lib.elrs_set_channels.argtypes = [ctypes.POINTER(ctypes.c_uint16)]
lib.elrs_set_channels.restype = None

lib.elrs_set_channel.argtypes = [ctypes.c_int, ctypes.c_uint16]
lib.elrs_set_channel.restype = None

lib.elrs_arm.argtypes = []
lib.elrs_arm.restype = None

lib.elrs_disarm.argtypes = []
lib.elrs_disarm.restype = None

lib.elrs_is_active.argtypes = []
lib.elrs_is_active.restype = ctypes.c_int

lib.elrs_set_linkstats_callback.argtypes = [ctypes.c_void_p]
lib.elrs_set_battery_callback.argtypes = [ctypes.c_void_p]
lib.elrs_set_gps_callback.argtypes = [ctypes.c_void_p]
lib.elrs_set_attitude_callback.argtypes = [ctypes.c_void_p]


class ELRSControl:
    """Python wrapper for ELRS joystick control"""
    
    # Channel indices
    ROLL = 0
    PITCH = 1
    THROTTLE = 2
    YAW = 3
    AUX1 = 4  # Usually ARM
    AUX2 = 5
    AUX3 = 6
    AUX4 = 7
    
    def __init__(self):
        self._initialized = False
        self._callbacks = {}  # Keep references to prevent garbage collection
        
    def init(self, port: str, baud_rate: int = 921600) -> bool:
        """Initialize connection to ELRS transmitter"""
        if self._initialized:
            raise RuntimeError("Already initialized")
            
        result = lib.elrs_init(port.encode('utf-8'), baud_rate)
        if result == 0:
            self._initialized = True
            return True
        else:
            error_msgs = {
                -1: "Already initialized",
                -2: "Failed to start link",
                -3: "Timeout waiting for link"
            }
            raise RuntimeError(f"Initialization failed: {error_msgs.get(result, 'Unknown error')}")
    
    def close(self):
        """Close connection and cleanup"""
        if self._initialized:
            lib.elrs_close()
            self._initialized = False
            self._callbacks.clear()
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
    
    def set_channels(self, channels: List[int]):
        """Set all 16 channels at once (values 0-1984, center=992)"""
        if not self._initialized:
            raise RuntimeError("Not initialized")
        if len(channels) != 16:
            raise ValueError("Must provide exactly 16 channel values")
            
        arr = (ctypes.c_uint16 * 16)(*channels)
        lib.elrs_set_channels(arr)
    
    def set_channel(self, channel: int, value: int):
        """Set a single channel (0-15, value 0-1984)"""
        if not self._initialized:
            raise RuntimeError("Not initialized")
        if not 0 <= channel <= 15:
            raise ValueError("Channel must be 0-15")
        if not 0 <= value <= 1984:
            raise ValueError("Value must be 0-1984")
            
        lib.elrs_set_channel(channel, value)
    
    def arm(self):
        """Arm the drone (sets AUX1 high)"""
        if not self._initialized:
            raise RuntimeError("Not initialized")
        lib.elrs_arm()
    
    def disarm(self):
        """Disarm the drone (sets AUX1 low)"""
        if not self._initialized:
            raise RuntimeError("Not initialized")
        lib.elrs_disarm()
    
    def is_active(self) -> bool:
        """Check if the link is active"""
        return lib.elrs_is_active() == 1
    
    def set_linkstats_callback(self, callback: Optional[Callable[[int, int, int, int], None]]):
        """Set callback for link statistics (rssi1, rssi2, lq%, snr)"""
        if callback:
            cb = LinkStatsCallback(callback)
            self._callbacks['linkstats'] = cb  # Keep reference
            lib.elrs_set_linkstats_callback(ctypes.cast(cb, ctypes.c_void_p))
        else:
            lib.elrs_set_linkstats_callback(None)
            self._callbacks.pop('linkstats', None)
    
    def set_battery_callback(self, callback: Optional[Callable[[float, float, float], None]]):
        """Set callback for battery telemetry (voltage, current, remaining%)"""
        if callback:
            cb = BatteryCallback(callback)
            self._callbacks['battery'] = cb
            lib.elrs_set_battery_callback(ctypes.cast(cb, ctypes.c_void_p))
        else:
            lib.elrs_set_battery_callback(None)
            self._callbacks.pop('battery', None)
    
    def set_gps_callback(self, callback: Optional[Callable[[float, float, int, int, float], None]]):
        """Set callback for GPS telemetry (lat, lon, alt_m, satellites, speed_m/s)"""
        if callback:
            cb = GPSCallback(callback)
            self._callbacks['gps'] = cb
            lib.elrs_set_gps_callback(ctypes.cast(cb, ctypes.c_void_p))
        else:
            lib.elrs_set_gps_callback(None)
            self._callbacks.pop('gps', None)
    
    def set_attitude_callback(self, callback: Optional[Callable[[float, float, float], None]]):
        """Set callback for attitude telemetry (pitch°, roll°, yaw°)"""
        if callback:
            cb = AttitudeCallback(callback)
            self._callbacks['attitude'] = cb
            lib.elrs_set_attitude_callback(ctypes.cast(cb, ctypes.c_void_p))
        else:
            lib.elrs_set_attitude_callback(None)
            self._callbacks.pop('attitude', None)
    
    # Helper methods for common operations
    def set_throttle(self, value: float):
        """Set throttle (0.0-1.0)"""
        self.set_channel(self.THROTTLE, int(value * 1984))
    
    def set_roll(self, value: float):
        """Set roll (-1.0 to 1.0)"""
        self.set_channel(self.ROLL, int((value + 1.0) / 2.0 * 1984))
    
    def set_pitch(self, value: float):
        """Set pitch (-1.0 to 1.0)"""
        self.set_channel(self.PITCH, int((value + 1.0) / 2.0 * 1984))
    
    def set_yaw(self, value: float):
        """Set yaw (-1.0 to 1.0)"""
        self.set_channel(self.YAW, int((value + 1.0) / 2.0 * 1984))
