# Testing Guide

## Interactive UI
1. Launch: `sidekick`
2. Ensure prompt line accepts input (including spaces)
3. Tab cycles **Ask / Edit / Plan**
4. Enter executes prompt when prompt line is selected
5. ↑/↓ selects Settings/Models/Help/Update/Quit
6. Enter opens selected menu
7. ←/Esc exits menus

## CLI
```bash
sidekick scan /path/to/project
```

## Reports
```bash
sidekick scan /path/to/project --format html --output report.html
```
