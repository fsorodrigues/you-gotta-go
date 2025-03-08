// UISwitch library can be found at https://github.com/MajicDesigns/MD_UISwitch
// MD_MAX72XX library can be found at https://github.com/MajicDesigns/MD_MAX72XX
#include <MD_Parola.h>
#include <MD_MAX72xx.h>
#include <SPI.h>

// Define the number of devices we have in the chain and the hardware interface
// NOTE: These pin numbers will probably not work with your hardware and may
// need to be adapted
#define HARDWARE_TYPE MD_MAX72XX::FC16_HW
#define MAX_DEVICES 4
#define CLK_PIN   13
#define DATA_PIN  11
#define CS_PIN    10

// HARDWARE SPI
MD_Parola P = MD_Parola(HARDWARE_TYPE, CS_PIN, MAX_DEVICES);

// Scrolling parameters
// uint8_t scrollSpeed = 25; // default frame delay value
// textEffect_t scrollEffect = PA_SCROLL_LEFT;
// textPosition_t scrollAlign = PA_LEFT;
// uint16_t scrollPause = 5000; // in milliseconds

// Global variables 
int ENCODING_VERSION = 1;
char MSGSTART = '<';
byte MSGSTART_BYTE = 0x3C;
char MSGEND = '>';
byte MSGEND_BYTE = 0x3E;
int HEADER_ENCODING_VERSION_OFFSET = 3;
int HEADER_LENGTH_OFFSET = 3;

const byte numBytes = 32;
byte receivedBytes[numBytes];
char receivedMsg[numBytes] = { "Init..." };
char displayMsg[numBytes] = { "" };
byte numReceived = 0;
bool newData = false;
bool newMsg = false;
char msg[numBytes];

void encodeMessage(char m[numBytes], char outgoingMsg[]) {
  int LEN = strlen(outgoingMsg);
  sprintf(m, "%c%3d%3d%s%c", MSGSTART, ENCODING_VERSION, LEN, outgoingMsg, MSGEND);
}

void readSerial() {
  static boolean recvInProgress = false;
  static byte ndx = 0;
  byte rb;

  while (Serial.available() > 0 && newData == false) {
    rb = Serial.read();

    if (recvInProgress == true) {
      if (rb != MSGEND_BYTE) {
        receivedBytes[ndx] = rb;
        ndx++;
        if (ndx >= numBytes) {
            ndx = numBytes - 1;
        }
      }
      else {
        receivedBytes[ndx] = '\0'; // terminate the string
        recvInProgress = false;
        numReceived = ndx;  // save the number for use when printing
        ndx = 0;
        newData = true;
        newMsg = true;
      }
    }

    else if (rb == MSGSTART_BYTE) {
      recvInProgress = true;
    }
  }
}

void getNewData() {
  if (newData == true) {
    for (byte ndx = HEADER_ENCODING_VERSION_OFFSET + HEADER_LENGTH_OFFSET; ndx < numBytes; ndx++) {
      receivedMsg[ndx] = (char *)receivedBytes[ndx];
    }

    newData = false;
  }
}

void setup()
{
  Serial.begin(57600);
  P.begin();
  P.displayText(displayMsg, PA_LEFT, 25, 5000, PA_SCROLL_LEFT, PA_SCROLL_LEFT);

  encodeMessage(msg, "ready");
  Serial.println(msg);
  delay(10);
  encodeMessage(msg, "start7932|23");
  Serial.println(msg);
}

void loop()
{
  if (P.displayAnimate())
  {
    if (true)
    {
      strcpy(displayMsg, receivedMsg);
      newMsg = false;
    }
    P.displayReset();
  }
 
  readSerial();
  getNewData();
}

