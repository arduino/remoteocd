# remoteocd

Flexible firmware flashing for the Arduino UNO Q Microcontroller.

`remoteocd` is a specialized utility designed to manage firmware deployment for the Arduino UNO Q board.
This tool acts as a versatile wrapper for OpenOCD (Open On-Chip Debugger), allowing you to flash a binary onto the MCU using one of three transparently handled modes:
- Local, by flashing from the UNO Q's MPU (Linux) environment.
- ADB over USB.
- SSH over a remote pc.

`remoteocd` is part of the `arduino:zephyr:unoq` platform.

## License

This project is licensed under the GPL3 license. See the LICENSE file for details.
