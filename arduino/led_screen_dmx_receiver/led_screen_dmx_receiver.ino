#include <Wire.h>
#include <DMXSerial.h>

#define SLAVE_ADDRESS 0x04
#define DMX_COUNT 12
char dmxs[DMX_COUNT+1];

// chan 0 : new dmx update
// chan 1 : display mode with inverse MSB for white dots (0 - on, 1 - off)
// chan 2 : color palette, if 0 then solid color
// chan 3 : palette angle, 0-360 mapped to 0-255
// chan 4 : palette offset
// chan 5 : solid color R
// chan 6 : solid color G
// chan 7 : solid color B
// chan 8 : solid color brightness
// chan 9 : lyric ID most significant byte
// chan 10: lyric ID cont
// chan 11: lyric ID least significant byte
// chan 12: lyric display progress 0-255

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

  for (int i = 0; i < DMX_COUNT; i++) {
    dmxs[i+1] = DMXSerial.read(i+1);
  }
  Wire.write(dmxs, DMX_COUNT+1);
}
