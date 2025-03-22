-- Variables to store the playback time and filename
local playback_time = 0
local filename = "unknown"
local username = "admin"
local password = "password"

-- URL of the REST API (replace with your server URL)
local api_url = "http://localhost:8080/timestamps"

local char_to_hex = function(c)
	return string.format("%%%02X", string.byte(c))
end

local function urlencode(url)
	if url == nil then
		return
	end
	url = url:gsub("\n", "\r\n")
	url = url:gsub("([^%w _%%%-%.~])", char_to_hex)
	url = url:gsub(" ", "+")
	return url
end

local hex_to_char = function(x)
	return string.char(tonumber(x, 16))
end

local urldecode = function(url)
	if url == nil then
		return
	end
	url = url:gsub("+", " ")
	url = url:gsub("%%(%x%x)", hex_to_char)
	return url
end

-- Function to capture the current playback time and filename during playback
function save_properties()
	playback_time = mp.get_property_number("time-pos", 0) -- Get current playback time in seconds
	filename = mp.get_property("filename", "unknown") -- Get the filename
end

-- Function to send the captured time and filename to the REST server on exit
function upload_time_on_exit()
	-- JSON payload containing both the playback time and filename
	local data = string.format('{"seconds": %f, "name": "%s"}', math.floor(playback_time), filename)
	-- POST request to upload the data
	local command =
		string.format("curl -s -X POST -H \"Content-Type: application/json\" -d '%s' -u '%s:%s' %s", data,username,password, api_url)

	-- Run the curl command and capture the result
	local handle = io.popen(command, "r")
	local result
	if handle ~= nil then
		result = handle:read("*a")
		if result ~= "" then
			-- Log the event (for debugging purposes)
			mp.msg.info("Uploaded playback time: " .. math.floor(playback_time) .. " seconds for file: " .. filename)
		else
			mp.msg.error("Couldn't upload playback time.")
		end
		handle:close()
	end

end

-- Function to fetch the saved timestamp from the server on file start
function check_server_for_timestamp()
	-- Get the filename of the current file
	filename = mp.get_property("filename", "unknown")

	-- API request to check if there is a timestamp for the current file
	local command = string.format('curl -s -u "%s:%s" "%s/%s"',username,password, api_url, urlencode(filename))
	-- Run the curl command and capture the result
	local handle = io.popen(command, "r")
	local result
	if handle ~= nil then
		result = handle:read("*all")
		handle:close()
	end
	if tonumber(result) then
		-- Parse the result (assuming the server returns the timestamp in seconds as plain text or JSON)
		local timestamp = tonumber(result)
		-- If the server returned a valid timestamp, seek to that position
		if timestamp and timestamp > 0 then
			mp.msg.info("Resuming playback at " .. math.floor(timestamp) .. " seconds for file: " .. filename)
			mp.set_property_number("time-pos", timestamp)
		else
			mp.msg.info("No saved timestamp found for file: " .. filename)
		end
	else
		mp.msg.error("Couldn't connect to the server or no timestamp found.") -- TODO fix this
	end
end

-- Hook to check the server for a saved timestamp when playback starts
mp.register_event("file-loaded", check_server_for_timestamp)

-- Hook the save_properties function to the "playback-restart" event to update values during playback
mp.register_event("playback-restart", save_properties)

-- Hook the upload function to the "shutdown" event, which triggers when MPV exits
mp.register_event("shutdown", upload_time_on_exit)

local save_timer = mp.add_periodic_timer(60, save_properties)
local function pause(name, paused)
	if paused then
		save_timer:stop()
		save_properties()
	else
		save_timer:resume()
	end
end
mp.observe_property("pause","bool",pause)
