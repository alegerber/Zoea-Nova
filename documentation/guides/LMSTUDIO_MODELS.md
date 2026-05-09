# Using LM Studio with Zoea Nova

Zoea Nova can use [LM Studio](https://lmstudio.ai) as a local provider via its
OpenAI-compatible HTTP server.

## Setup

1. Install LM Studio and load a tool-calling-capable model (recommended:
   Qwen3.5-9B, Llama-3.1-8B-Instruct).
2. In LM Studio, open **Developer → Local Server** and click **Start Server**.
   Default URL: `http://localhost:1234/v1`.
3. Add to `~/.zoea-nova/config.toml`:

   ```toml
   [providers.lm-qwen]
   type = "lmstudio"
   endpoint = "http://localhost:1234/v1"
   model = "qwen/qwen3.5-9b"
   temperature = 0.7
   ```

   The `model` value must match the API identifier shown in LM Studio's server
   panel — often the Hugging-Face `<publisher>/<slug>` form.

4. Optionally set `default_provider = "lm-qwen"` under `[swarm]`.
5. Restart Zoea Nova.

## Verifying

After startup, the log line `Providers initialized providers=N` should include
the LM Studio entry. In the TUI, press `c` to configure a mysis and enter the
provider name (e.g. `lm-qwen`).

## Troubleshooting

- **`provider not found`**: type is misspelled or the endpoint is invalid.
- **`connection refused`**: LM Studio's local server is not running, or the
  port differs from the one in `endpoint`.
- **HTTP 404 from `/v1/chat/completions`**: the `model` slug doesn't match
  exactly what LM Studio's server panel shows.
- **No tool calls**: the loaded model does not support tool calling — switch
  to a tool-capable model in LM Studio.

## Notes

- LM Studio does not require authentication.
- Streaming is supported.
- Reasoning models (e.g. DeepSeek-R1 distillates) emit reasoning inside
  `content`; the adapter does not do separate reasoning extraction.
