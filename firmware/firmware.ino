#include <ESP8266WiFi.h>
//#include <WiFiClient.h>
#include <ESP8266WebServer.h>
#include <ESP8266HTTPClient.h>
#include <SoftwareSerial.h>

#define STASSID "EL-BETHEL"
#define STAPSK "150230111995"

#define LEDPIN D0

const char *ssid = STASSID;
const char *password = STAPSK;

ESP8266WebServer server(80);
SoftwareSerial arduinoSerial(D1, D2);

String prometheusMetrics = "";

int PIR_PIN = D3;

void handleMetrics()
{
    digitalWrite(LEDPIN, LOW);
    arduinoWriteLine("metrics");
    String prometheusMetrics = arduinoReadLine("METRICS");
    server.send(200, "text/plain", prometheusMetrics);
    digitalWrite(LEDPIN, HIGH);
}

void handleNotFound()
{
    digitalWrite(LEDPIN, LOW);
    String message = "File Not Found\n\n";
    message += "URI: ";
    message += server.uri();
    message += "\nMethod: ";
    message += (server.method() == HTTP_GET) ? "GET" : "POST";
    message += "\nArguments: ";
    message += server.args();
    message += "\n";
    for (uint8_t i = 0; i < server.args(); i++)
    {
        message += " " + server.argName(i) + ": " + server.arg(i) + "\n";
    }
    server.send(404, "text/plain", message);
    digitalWrite(LEDPIN, HIGH);
}

void handleArduino()
{
    digitalWrite(LEDPIN, LOW);
    int args = server.args();
    if (args > 0)
    {
        String message = "";
        for (uint8_t i = 0; i < args; i++)
        {
            message += server.argName(i) + "=" + server.arg(i) + "$";
        }
        // Send all at once - otherwise it shows up on arduino garbled
        arduinoWriteLine(message);
        server.send(200, "text/plain", F("OK"));
    }
    else
    {
        server.send(400, "text/plain", F("KO"));
    }
    digitalWrite(LEDPIN, HIGH);
}

void setup()
{
    pinMode(LEDPIN, OUTPUT);
    digitalWrite(LEDPIN, LOW);
    delay(3000);
    digitalWrite(LEDPIN, HIGH);
    Serial.begin(115200);
    arduinoSerial.begin(14400);

    Serial.println();
    Serial.print("Configuring access point...");
    WiFi.mode(WIFI_STA);
    WiFi.begin(ssid, password);
    Serial.println();

    // Wait for connection
    Serial.print("Connecting.");
    while (WiFi.status() != WL_CONNECTED)
    {
        delay(500);
        Serial.print(".");
    }
    Serial.println();
    Serial.print("Connected to ");
    Serial.println(ssid);
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
    Serial.print("DNS address: ");
    Serial.println(WiFi.dnsIP());

    server.on("/inline", []() {
        server.send(200, "text/plain", "this works as well");
    });

    server.on("/metrics", handleMetrics);

    server.on("/arduino", handleArduino);

    server.onNotFound(handleNotFound);

    server.begin();
    Serial.println("HTTP server started");

    // temporary code for PIR sensor
    pinMode(PIR_PIN, INPUT);
}

void arduinoWriteLine(String msg)
{
    Serial.println("Sending message: " + msg);
    arduinoSerial.print(msg + "\n");
}

String arduinoReadLine(String prefix)
{
    //    String msg = "";
    //    char ch;
    //    while (ch != '\n') {
    //        ch = arduinoSerial.read();
    //        msg += ch;
    //    } ;
    unsigned long startTime = millis();
    while (true)
    {
        while (arduinoSerial.available() <= 0)
        {
            if (millis() - startTime > 1000)
            {
                return "";
            }
        }

        String msg = arduinoSerial.readStringUntil('\n');
        // Serial.println("Received message: " + msg);
        if (!msg.startsWith(prefix + ": "))
        {
            continue;
        }
        msg = msg.substring(prefix.length() + 2);
        msg.replace('$', '\n');
        return msg;
    }
    return "";
}

void loop()
{
    server.handleClient();
    handlePIR();
}

// temporary code for PIR sensor
int pir_value = -1;

void handlePIR()
{
    int new_pir_value = digitalRead(PIR_PIN);
    if (new_pir_value == pir_value)
        return;
    pir_value = new_pir_value;
    sendEvent(PIR_PIN, pir_value);
}

void sendEvent(byte pin, bool state)
{
    HTTPClient http;
    String req_string;
    String ip = "192.168.1.5";
    req_string = "http://";
    req_string += ip;
    req_string += "/api/webhook";
    Serial.println(req_string);
    http.begin(req_string);
    http.addHeader("Content-Type", "text/plain");

    String put_string;
    put_string = "pin=";
    put_string += pin;
    put_string += "&state=";
    put_string += state;
    Serial.println(put_string);

    int httpResponseCode = http.PUT(put_string);

    if (httpResponseCode > 0)
    {
        // String response = http.getString();
        Serial.println(httpResponseCode);
        // Serial.println(response);
    }
    else
    {
        Serial.print("Error on sending PUT Request: ");
        Serial.println(httpResponseCode);
    }
    http.end();
}
