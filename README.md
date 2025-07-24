# ELRS Joystick Control

## Building

```bash
go build -tags static -trimpath --ldflags '-s -w' -o elrs-joystick-control ./cmd/elrs-joystick-control/.
```

## How It Works

The application reads the raw inputs from one or more USB gamepad devices. It takes these
inputs, converts them to Crossfire format (CRSF), and sends them to an RC Transmitter (TX) module.

The TX then sends the control signals over air to the drone.  Both the USB control devices and the
RC Transmitter module must be connected to the same computer where the application is running on.

## How the application talks to the ELRS Transmitter

ELRS TX modules have an I/O pin that is used for receiving radio inputs.

The transmitter module does not really care who is sending data on that pin. It could be an actual device like a
Radio Master TX16S, or it could be this application.

This application uses a serial port to send data to the ELRS TX, and in doing so, pretends to be an RC radio.

## Connecting to the ELRS Transmitter via USB

Some ELRS transmitters have a USB port that is used for flashing firmware. (otherwise, need to use FTDI adapter)
This same USB port can be reconfigured to work as the CRSF I/O pin.

First, download STM32 Virtual COM Port driver, from the [ST Electronics website](https://www.st.com/en/development-tools/stsw-stm32102.html)

Then, access the module's /hardware.html page, and change the CRSF RX/TX pin values.

The correct values to use here depend on the module you have.
For example, in my case, with the BetaFPV 1W Micro module, I had to use pins 3 and 1 so that the
ELRS firmware would treat the USB port as if it was the CRSF serial port.

You can usually tell which pin values to use by looking at the ELRS Backpack/Logging configuration
(in the same hardware.html page).

The ELRS Backpack/Logging section is configured by default to use the USB RX/TX pins.
So, copy+paste these values and disable the backpack functionality.

Finally, you may need to put your ELRS TX in "Firmware Upgrade" mode for this approach to work.
This is done using the DIP switch on the back of the module. The exact position of the DIP switch varies
from module to module. See the ELRS documentation to determine the proper method for putting the module in "Firmware upgrade" mode.

## How to power the ELRS transmitter module

There are a few ways you can power the transmitter module without connecting it to the JR bay of an existing radio.

  * **USB Power** - First, you can power the ELRS transmitter using the USB connector (if it has one). The RF output power will be
limited when using USB power. It's very likely that you will not be able to go over 100 milli-watts of RF output power.
That's still plenty of power for most flying. But beware, if you set the transmitter's RF output too high, it may 
exceed the power supply from the USB connection. This can cause the module to brown-out, and reboot itself. 
It will keep rebooting, and shutting down. If this happens to you, you will need to connect the module to a higher wattage power supply, and revert the settings.

  * **XT30 DC input** - The second approach is to use the module's XT30 DC input (if it has one). But beware, some modules may not have protection
to isolate the XT30 DC input from rest of the circuitry. Early versions of Radio-Master ELRS Ranger 
transmitters had this issue. Some pilots damaged their radios when they connected the XT30 input at the same time they had the module 
connected to the JR bay of the radio. So, don't do that.

  * **JR Bay VCC / GND pins** - The third and final approach is to use the JR bay `VCC` / `GND` pins. Most ELRS transmitter modules accept between 5V and 12V across the
`VCC` / `GND` pins. You can connect a 2S LiPo battery directly to the these pins. ELRS transmitter modules have an internal voltage
regulator, so it should be safe.

## How to use the gRPC service

In order to use the gRPC service, you will need a gRPC client. There are a few of those out there like Postman, or [GRPC-UI](https://github.com/fullstorydev/grpcui/releases).

Here is are some instructions for GRPC-UI

1. Download and extract the GRPC-UI binary from their [GitHub releases](https://github.com/fullstorydev/grpcui/releases).
    * Put the `grpcui` binary somewhere in your path
2. Start the **elrs_joystick_control** application (by default it listens on port 10000)
    ```shell
    $ elrs_joystick_control
     gRPC server listenting on port 10000
    ```
3. Start GRPC-UI like this
    ```shell
    $ grpcui -plaintext localhost:10000
     gRPC Web UI available at http://127.0.0.1:53885/
    ```

From GRPC-UI, you can call the methods exposed by the application's gRPC service. The following main methods are available:

* **setConfig** - Receives (and validates) a JSON file containing the full configuration, and stores it in memory
* **getConfig** - Retrieves the full configuration from memory, and sends it as a JSON file

* **startLink** - starts the link with the RF transmitter
* **stopMixer** - stops the link with the RF transmitter

* **startHttp** - Starts the Web-UI HTTP server
* **stopHTTP** - Stops the Web-UI HTTP server

* **getGamepads** - Returns a list of raw input devices connected (joysticks, gamepads, etc)
* **getTransmitters** - Returns a list of available serial ports

There are also a few other data streaming methods available:

* **getEvalStream** - Starts a data stream with the values for all inputs/outputs as they are config is evaluated live
* **getTransmitterStream** - Starts a data stream with the values of all 16 channels as they are received live by the RF transmitter.
* **getGamepadStream** - Starts a data stream with the values of all axes, and buttons as they are output by a gamepad
* **getTelemetryStream** - Starts a data stream with the values of all telemetry frames that are output by the ELRS TX
* **getLinkStream** - Starts a data stream with values for link stats such as count of sent/received frames, and errors.

