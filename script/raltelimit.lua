-- KEYS[1] = bucket key
-- ARGV[1] = capacity
-- ARGV[2] = rate (tokens per second)
-- ARGV[3] = now (unix timestamp, seconds)

local bucket = redis.call("HMGET", KEYS[1], "tokens", "ts")

local tokens = tonumber(bucket[1])
local ts = tonumber(bucket[2])

if tokens == nil then
    tokens = tonumber(ARGV[1])
    ts = tonumber(ARGV[3])
end

local delta = math.max(0, tonumber(ARGV[3]) - ts)
local filled = math.min(tonumber(ARGV[1]), tokens + delta * tonumber(ARGV[2]))

if filled < 1 then
    redis.call("HMSET", KEYS[1], "tokens", filled, "ts", ARGV[3])
    redis.call("EXPIRE", KEYS[1], 2)
    return 0
end

redis.call("HMSET", KEYS[1], "tokens", filled - 1, "ts", ARGV[3])
redis.call("EXPIRE", KEYS[1], 2)
return 1
