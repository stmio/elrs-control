# ELRS Control

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
