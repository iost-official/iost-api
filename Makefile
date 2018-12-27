GO = go

TARGET = iost-api

all:$(TARGET)

iost-api: main.go
	$(GO) build -o $@ $^

clean:
	rm -rf $(TARGET)