#ifndef CALLBACKS_H
#define CALLBACKS_H

void call_linkstats_callback(void* fn, int rssi1, int rssi2, unsigned int lq, int snr);
void call_battery_callback(void* fn, float voltage, float current, float remaining);
void call_gps_callback(void* fn, float lat, float lon, int alt, unsigned int sats, float speed);
void call_attitude_callback(void* fn, float pitch, float roll, float yaw);

#endif
