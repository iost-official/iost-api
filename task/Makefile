GO = go

TARGET = iost-api-task

all:$(TARGET)

iost-api-task: task.go
	$(GO) build -o $@ $^

clean:
	rm -rf $(TARGET)
