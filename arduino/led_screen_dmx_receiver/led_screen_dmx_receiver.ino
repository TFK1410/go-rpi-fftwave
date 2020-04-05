#include <Wire.h>
#include <DMXSerial.h>

#define SLAVE_ADDRESS 0x04
char dmxs[8];

void setup()
{
  Wire.begin(SLAVE_ADDRESS); // join i2c bus
  Wire.onRequest(requestEvent); // register event
  DMXSerial.init(DMXReceiver);
}

void loop()
{
  delay(10000);
}

void requestEvent() {
  if (DMXSerial.dataUpdated()){
    dmxs[0] = 255;
    DMXSerial.resetUpdated();
  }
  else
    dmxs[0] = 0;
  //LSB Measure
  dmxs[1] = DMXSerial.read(1);
  //MSB Measure
  dmxs[2] = DMXSerial.read(2);
  //LSB Track ID
  dmxs[3] = DMXSerial.read(3);
  //MSB Track ID
  dmxs[4] = DMXSerial.read(4);
  //Color R
  dmxs[5] = DMXSerial.read(5);
  //Color G
  dmxs[6] = DMXSerial.read(6);
  //Color B
  dmxs[7] = DMXSerial.read(7);
  Wire.write(dmxs, 8);
}
