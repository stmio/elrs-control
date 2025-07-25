#include "callbacks.h"

#include <stdint.h>

typedef void (*linkstats_callback_t)(int32_t rssi1, int32_t rssi2, uint32_t lq,
                                     int32_t snr);
typedef void (*battery_callback_t)(float voltage, float current,
                                   float remaining);
typedef void (*gps_callback_t)(float lat, float lon, int32_t alt, uint32_t sats,
                               float speed);
typedef void (*attitude_callback_t)(float pitch, float roll, float yaw);

void call_linkstats_callback(void *fn, int rssi1, int rssi2, unsigned int lq,
                             int snr) {
  ((linkstats_callback_t)fn)(rssi1, rssi2, lq, snr);
}

void call_battery_callback(void *fn, float voltage, float current,
                           float remaining) {
  ((battery_callback_t)fn)(voltage, current, remaining);
}

void call_gps_callback(void *fn, float lat, float lon, int alt,
                       unsigned int sats, float speed) {
  ((gps_callback_t)fn)(lat, lon, alt, sats, speed);
}

void call_attitude_callback(void *fn, float pitch, float roll, float yaw) {
  ((attitude_callback_t)fn)(pitch, roll, yaw);
}
