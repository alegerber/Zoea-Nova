# Using LM Studio with Zoea Nova

Zoea Nova can use [LM Studio](https://lmstudio.ai) as a local provider via its
OpenAI-compatible HTTP server.

## Setup

1. Install LM Studio and load a tool-calling-capable model (recommended:
   Qwen3.5-9B, Llama-3.1-8B-Instruct). The display name shown in LM Studio's
   "My Models" panel may differ from the API identifier you'll need below
   (e.g. display "Qwen3.5 9B" → API identifier `qwen/qwen3.5-9b`); always use
   the API identifier as it appears in the **Local Server** panel for the
   `model` field. Adjust the example slug below to match your actual loaded model.
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

After restarting Zoea Nova, the LM Studio entry should be available in the
TUI's mysis-creation flow. Press `n` to create a new mysis: enter a name,
then the provider (e.g. `lm-qwen`), then a model (e.g. `qwen/qwen3.5-9b`).

If a configured provider is silently missing, run Zoea Nova with debug
logging enabled and look for `"Providers initialized"` to see what was
registered, plus any `"skipping provider"` warnings indicating registration
failures.

## Troubleshooting

- **`provider not found` at runtime**: the provider name in the mysis config does not match any `[providers.*]` entry in `~/.zoea-nova/config.toml`. Check spelling.
- **Provider silently missing at startup**: if `type` is mistyped, `initProviders` logs `"skipping provider — no handler registered for type"` (zerolog warn level) and skips registration. Check the log on startup if a configured provider doesn't appear in the TUI.
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
