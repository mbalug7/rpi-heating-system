# Heating Pump Controller App

The Heating Pump Controller App is designed to run on a Raspberry Pi with kernel version 5.9 or higher. Its primary purpose is to control relays connected to heating pumps. The app is built to integrate with Home Assistant, where thermostats take charge of controlling the relays for the heating system.

## Features

- **Pump Control:** The app allows controlling the heating pumps connected to relays. It can switch the pumps ON or OFF as directed by the Home Assistant thermostats.

- **Integration with Home Assistant:** The app seamlessly integrates with Home Assistant, enabling a smooth communication channel between the thermostats and the heating pumps.

- **Temperature Sensing:** The app includes support for temperature sensors strategically placed to measure the temperature in the following locations:

  - **Buffer Tank:** Sensors are used to measure the temperature in the big buffer tank.

  - **Inlet and Outlet Temperature:** Sensors are also used to measure the inlet and outlet temperatures of the water in the heating system. This data is crucial for the Home Assistant thermostats to make informed decisions for heating control.

## Requirements

- Raspberry Pi with kernel version 5.9 or higher
- Relays for controlling the heating pumps
- Home Assistant installation with thermostats for heating control
- One-wire temperature sensors (DS18B20) for temperature measurement

## How to Use

1. Install the app on the Raspberry Pi with the appropriate kernel version.
2. Connect the relays to the heating pumps as needed.
3. Set up Home Assistant and configure the thermostats for heating control.
4. Run the Heating Pump Controller App, and it will automatically integrate with Home Assistant to receive control commands.

## Installation

To install the app on the Raspberry Pi, follow these steps:

1. Clone the repository to the Raspberry Pi.
2. Ensure the kernel version is 5.9 or higher on the Raspberry Pi.
3. Build the app using `make build`.
4. Transfer app to Raspberry Pi using `make transfer HOST=pi@ip_address`
5. Transfer config to Raspberry Pi, or create it manually in the same directory where binary is placed. 
6. Run the app using `./heating_pump_controller`.

## Contributing

Contributions to the Heating Pump Controller App are welcome! If you find any issues or have ideas for enhancements, please open an issue or submit a pull request.

