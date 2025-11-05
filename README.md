# remoteocd

Flexible firmware flashing for the Arduino UNO Q Microcontroller.

`remoteocd` is a specialized utility designed to manage firmware deployment for the Arduino UNO Q board.
This tool acts as a versatile wrapper for OpenOCD (Open On-Chip Debugger), allowing you to flash a binary onto the MCU using one of three transparently handled modes:

- Local, by flashing from the UNO Q's MPU (Linux) environment.
- ADB over USB.
- SSH over a remote pc.

`remoteocd` is part of the `arduino:zephyr:unoq` platform.

### Uploading with remoteocd 
Uploading the compiled binary to the microcontroller is handled by remoteocd. Instead of relying on hard-coded flash commands, it accepts an OpenOCD configuration file. This allows you to tailor the upload script to specific debugging or flashing requirements.
remoteocd is automatically installed as a tool dependency of the [Arduino UNO Q Board platform](https://github.com/arduino/ArduinoCore-zephyr).

To upload your compiled sketch, use the following Arduino CLI command:

```bash
arduino-cli compile -b arduino:zephyr:unoq

arduino-cli upload -b arduino:zephyr:unoq
```

## License

This project is licensed under the GPL3 license. See the LICENSE file for details.

