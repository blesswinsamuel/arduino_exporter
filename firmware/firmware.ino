// REQUIRES the following Arduino libraries:
// - DHT Sensor Library: https://github.com/adafruit/DHT-sensor-library
// - Adafruit Unified Sensor Lib: https://github.com/adafruit/Adafruit_Sensor
#include <DHT.h>
#include <SoftwareSerial.h>
#include <LiquidCrystal.h>

#define DHTPIN 4
#define LEDPIN 5
#define BUZZERPIN 6
#define LDRPIN A0

// Connect pin 1 (on the left) of the sensor to +5V
// NOTE: If using a board with 3.3V logic like an Arduino Due connect pin 1
// to 3.3V instead of 5V!
// Connect pin 2 of the sensor to whatever your DHTPIN is
// Connect pin 4 (on the right) of the sensor to GROUND
// Connect a 10K resistor from pin 2 (data) to pin 1 (power) of the sensor

DHT dht(DHTPIN, DHT11);
LiquidCrystal lcd(12, 11, 7, 8, 9, 10);

void setup()
{
    Serial.begin(9600); // communication with the host computer
    dht.begin();

    lcd.begin(16, 2); // set up the LCD's number of columns and rows
    lcd.print("hello!");

    pinMode(LEDPIN, OUTPUT);
    pinMode(BUZZERPIN, OUTPUT);
}

byte ledState = LOW;
byte buzzerState = LOW;
String lcd1Value = "Hello!";
String lcd2Value = "------";
int ledDuration = -1;
int buzzerDuration = -1;

float humidity = -1;
float temperature = -1;
float heatIndex = -1;
float ldrValue = -1;

unsigned long currentMillis = 0;        // stores the value of millis() in each iteration of loop()
unsigned long previousLedMillis = 0;    // will store last time the LED was updated
unsigned long previousBuzzerMillis = 0; // will store last time the Buzzer was updated
unsigned long previousDhtReadMillis = 0;

void loop()
{
    currentMillis = millis();
    updateLedState();
    updateBuzzerState();

    listenSerial();

    writeState();
}

void listenSerial()
{
    // listen for communication from the ESP8266 and then write it to the serial monitor
    if (Serial.available())
    {
        String args = Serial.readStringUntil('\n');
        if (args == "metrics")
        {
            readLdr(); // Take LDR reading before flashing the LED
            digitalWrite(LEDPIN, HIGH);
            readDht();
            String metrics = createMetricsString();
            metrics.replace("\n", "$");
            Serial.println("METRICS: " + metrics);
        }
        else
        {
            int sep = args.indexOf("=");
            String argName = args.substring(0, sep);
            String argValueStr = args.substring(sep + 1);
            Serial.println("LOG: " + argName + "=" + argValueStr + "$");
            if (argName == "led")
            {
                int argValue = argValueStr.toInt();
                previousLedMillis = currentMillis;
                ledDuration = argValue;
                ledState = HIGH;
            }
            else if (argName == "buzzer")
            {
                int argValue = argValueStr.toInt();
                previousBuzzerMillis = currentMillis;
                buzzerDuration = argValue;
                buzzerState = HIGH;
            }
            else if (argName == "lcd1")
            {
                lcd.clear();
                lcd1Value = argValueStr;
            }
            else if (argName == "lcd2")
            {
                lcd.clear();
                lcd2Value = argValueStr;
            }
        }
    }
}

void updateLedState()
{
    if (ledState == HIGH)
    {
        if (currentMillis - previousLedMillis >= ledDuration)
        {
            ledState = LOW;
            previousLedMillis = currentMillis;
        }
    }
}

void updateBuzzerState()
{
    if (buzzerState == HIGH)
    {
        if (currentMillis - previousBuzzerMillis >= buzzerDuration)
        {
            buzzerState = LOW;
            previousBuzzerMillis = currentMillis;
        }
    }
}

void writeState()
{
    digitalWrite(LEDPIN, ledState);
    digitalWrite(BUZZERPIN, buzzerState);
    lcd.setCursor(0, 0);  // set the cursor to column 0, line 0
    lcd.print(lcd1Value); // Print a message to the LCD.
    lcd.setCursor(0, 1);  // set the cursor to column 0, line 1
    lcd.print(lcd2Value); // Print a message to the LCD.
}

void readDht()
{
    const int readDuration = 2000;
    if (currentMillis - previousDhtReadMillis >= readDuration)
    {
        previousDhtReadMillis = currentMillis;

        // Reading temperature or humidity takes about 250 milliseconds!
        // Sensor readings may also be up to 2 seconds 'old' (its a very slow sensor)
        float h = dht.readHumidity();
        // Read temperature as Celsius (the default)
        float t = dht.readTemperature();

        // Check if any reads failed and exit early (to try again).
        if (isnan(h) || isnan(t))
        {
            Serial.println(F("LOG: Failed to read from DHT sensor!"));
            return;
        }
        temperature = t;
        humidity = h;
        heatIndex = dht.computeHeatIndex(temperature, humidity, false);
    }
}

void readLdr()
{
    ldrValue = analogRead(LDRPIN);
}

String createMetricsString()
{
    String metrics = "";
    if (temperature != -1)
    {
        metrics += "uno_dht_temperature " + String(temperature) + "\n";
    }
    if (humidity != -1)
    {
        metrics += "uno_dht_humidity " + String(humidity) + "\n";
    }
    if (heatIndex != -1)
    {
        metrics += "uno_dht_heat_index " + String(heatIndex) + "\n";
    }
    if (ldrValue != -1)
    {
        metrics += "uno_ldr_value " + String(ldrValue) + "\n";
    }
    return metrics;
}

// arduino-cli compile --fqbn arduino:avr:uno --upload --port /dev/cu.usbmodem1421301 arduino/Metrics
