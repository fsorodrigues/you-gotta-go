#include <LiquidCrystal.h>

// https://www.arduino.cc/reference/en/libraries/liquidcrystal/liquidcrystal/
// LiquidCrystal(rs, enable, d4, d5, d6, d7)
LiquidCrystal lcd(12, 11, 5, 4, 3, 2);

const byte numBytes = 32;
byte receivedBytes[numBytes];
byte displayBytes[numBytes];
byte numReceived = 0;
boolean newData = false;
boolean newMsg = false;

void setup() {
  Serial.begin(9600);
  Serial.println("<Arduino is ready>");

  // set up the LCD's number of columns and rows
  lcd.begin(16, 2);
  Serial.println("<LCD ready>");
}

void loop() {
  recvBytesWithStartEndMarkers();
  getNewData();
  printData();
}

void recvBytesWithStartEndMarkers() {
  static boolean recvInProgress = false;
  static byte ndx = 0;
  byte startMarker = 0x3C;
  byte endMarker = 0x3E;
  byte rb;
  

  while (Serial.available() > 0 && newData == false) {
    rb = Serial.read();

    if (recvInProgress == true) {
      if (rb != endMarker) {
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

    else if (rb == startMarker) {
      recvInProgress = true;
    }
  }
}

void getNewData() {
  if (newData == true) {
    Serial.print("Incoming new data... ");
    Serial.print(numReceived);
    Serial.print(" bytes received");
    Serial.println();
  
    for (byte ndx = 0; ndx < numBytes; ndx++) {
      displayBytes[ndx] = receivedBytes[ndx];
    }

    newData = false;
  }
}

void printData() {
  char* msg = (char *)displayBytes;
  if (newMsg == true) {
    lcd.clear();
    lcd.setCursor(0, 0);
    lcd.print(msg);

    newMsg = false;
  }
  // if (strncmp("23:", msg, 3) == 0) {
  //   digitalWrite(LED_BUILTIN, HIGH);  // turn the LED on (HIGH is the voltage level)
  //   delay(500);                      // wait for a second
  //   digitalWrite(LED_BUILTIN, LOW);   // turn the LED off by making the voltage LOW
  //   delay(500);    
  // }
}

