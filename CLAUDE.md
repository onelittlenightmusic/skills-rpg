# skill-rpg CLAUDE.md

## Documentation

### STAGES.md
Records the learning theme, doors, devices, keys, clear conditions, and goal steps for each stage.
Whenever a stage is added or changed, `STAGES.md` must also be updated.

---

## Development Notes

### Server reset required after changing stage YAMLs

After editing `stages/*.yaml` and restarting the server, as long as `~/.mywant-rpg/current.yaml` exists, the server will load from that file (the stage YAMLs will not be re-read).

To apply new stage content to the server, always run a reset after restarting:

```bash
curl -s -X POST http://localhost:7100/api/v1/reset
```

Or via the MCP tool.

**Future improvement idea**: On server startup, compare the mtime of stage YAMLs against current.yaml, and auto-reboot if the YAMLs are newer.
