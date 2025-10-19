# How to set up the project:

1. **Clone the project**:
   ```bash
   git clone https://github.com/fastboyz/safelyYouCodingChallennge.git
    ```
2. **Navigate to the project folder**:
3. ```bash
   cd safelyYouCodingChallenge
   ```
4. **Install all dependencies**:
   ```bash
   go mod tidy
   ```
5. **Run the application**:
   ```bash
   go run main.go
   ```
   The server will start listening on port 6733
6. Run the device simulator:
   ```bash
   ./device_simulator(.exe on Windows)
   ```
   This will simulate device data being sent to the server.
7. Get the result.txt file in the same directory as the device simulator.

# Project Structure

- `main.go`: The main entry point of the application.
- `./api`: Contains The generated API code using OpenAPI specifications and the Server implementation.
- `./models`: Contains the device data models.
- `./services`: Contains the logic for processing device data.

# Questions

## How long did you spend working on the problem? What did you find to be the most difficult part?

I spent approximatively 4 hours working on the problem.
The most difficult part was designing the data processing logic to ensure it met the requirements
while maintaining code clarity and efficiency.

## How would you modify your data model or code to account for more kinds of metrics?

To introduce a new metric, I would:

1. Add the new metric model in the models `models/devices.go`.
2. Add the new metric in a field of the `Device` struct.
3. Add The new endpoint handler in `api/server_impl.com` to receive the new metric data.
4. Add the processing logic for the new metric in `services/device_communication_service.go`.

Example: If we wanted to add a new metric for the CPU, containing fields like `usage_percentage` and `temperature`.

1. Define a new struct in `models/devices.go`:
   ```go
   type CPUMetric struct {
       UsagePercentage float64 `json:"usage_percentage"`
       Temperature     float64 `json:"temperature"`
   }
   ```
2. Add a field in the `Device` struct:
    ```go
    type Device struct {
    DeviceID       string `json:"device_id"`
    Heartbeats     []*Heartbeat
    Stats          []*Stat
    SumUploadTime  int
    StatCount      int
    HeartbeatCount int
    FirstHeartbeat time.Time
    LastHeartbeat  time.Time
    CPUMetrics     []*CPUMetric
    }
    ```   
3. Add a new endpoint handler in `api/server_impl.go`:
    ```go
    func (s *ServerImpl) PostCPUMetric(w http.ResponseWriter, r *http.Request, deviceId DeviceIDPathParam) error {
        var req CpuMetricRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		            errResponse := Error{Msg: "error decoding body"}
		            log.Err(err).Msg("error decoding body")
		            w.WriteHeader(http.StatusBadRequest)
		            _ = json.NewEncoder(w).Encode(errResponse)
		            return
	        }
        return s.deviceService.ProcessCPUMetric(deviceID, req)
    }

    ```
4. Implement the processing logic in `services/device_communication_service.go`:
    ```go
    func (s *DeviceCommunicationService) ProcessCPUMetric(deviceID string, request CpuMetricRequest) {
            device, exists := s.devices[deviceID]
            if !exists {
                device = &models.Device{
                    DeviceID: deviceID,
                }
                s.devices[deviceID] = device
            }
            cpuMetric := &models.CPUMetric{
                UsagePercentage: request.UsagePercentage,
                Temperature:     request.Temperature,
            }
    
            device.CPUMetrics = append(device.CPUMetrics, cpuMetric)
    }
   ```
   

## Discuss your solution’s runtime complexity

1. CSV Ingestion: O(n) - Where n is the number of lines in the CSV file. Each line is read and processed once.
2. Data Processing: O(1) - Each device's data is processed in constant time as we are only updating the latest values. This is done by using a map to retrieve the device by its ID.
3. Get Stats for a Device: O(1) - Retrieving the latest stats for a device is done in constant time using a map. This is achieved by keeping track of sums and counts outside arrays to avoid O(n) operations when computing averages.
4. Overall Complexity: O(n) - The overall complexity is dominated by the CSV ingestion step.