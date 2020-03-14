#include <Wire.h>
#include <DMXSerial.h>

#define SLAVE_ADDRESS 0x04
char dmxs[4];

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
  dmxs[1] = DMXSerial.read(1);
  dmxs[2] = DMXSerial.read(2);
  dmxs[3] = DMXSerial.read(3);
  Wire.write(dmxs, 4);
}
