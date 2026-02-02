# Flow Redesign - Using Raw LLM Output

## Problem
The previous implementation forced the AI to format its findings in a specific structured format (`ISSUE:/LINE:/RECOMMENDATION:`), then used regex parsing to extract the data. This rigid approach caused data loss when the LLM didn't follow the exact format, resulting in issues being found but not appearing in reports.

## Solution
Redesigned the data flow to **primarily use raw LLM output** instead of forcing structured parsing:

### Key Changes

1. **Updated Data Structure** (`internal/scanner/scanner.go`)
   ```go
   type ScanResult struct {
       FilePath     string
       RawFindings  string          // The raw AI output
       HasIssues    bool           // Whether issues were found
       Issues       []SecurityIssue // Deprecated: kept for compatibility
   }
   ```

2. **Simplified Stage 3** (`internal/scanner/security.go`)
   - Removed `formatAsStructured()` which forced rigid formatting
   - `formatAsText()` now asks AI to format naturally for terminal (plain text)
   - `formatAsHTML()` asks AI to create clean HTML directly
   - No more regex parsing that loses data

3. **Updated Display Functions**
   - `cmd/scan.go::displayResults()` - Shows raw findings with orange headers
   - `internal/interactive/scan.go::displayResults()` - Same for interactive mode
   - Both now count files scanned vs files with findings (instead of issue counts)

4. **Updated HTML Report Generator** (`internal/report/html.go`)
   - Removed severity counters (CRITICAL/HIGH/MEDIUM/LOW)
   - Shows "Files Scanned" and "Files With Findings" instead
   - Displays raw LLM output as HTML using `safeHTML` template function
   - Simplified summary cards

5. **Fixed Result Collection** (`internal/scanner/scanner.go`)
   - Now collects ALL scan results (not just those with issues)
   - This fixes the "Files scanned: 0" bug

## Benefits

✅ **No data loss** - All LLM findings appear in reports  
✅ **More flexibility** - AI can express findings naturally  
✅ **Better formatting** - AI creates clean, readable output for each format (text/HTML)  
✅ **Simpler code** - Removed complex regex parsing logic  
✅ **Future-proof** - Easy to add new output formats

## Testing

Tested with `examples/vulnerable_code.go`:

**Text format:**
```bash
./sidekick scan examples/vulnerable_code.go --format text
```
Shows all 4+ vulnerabilities with proper formatting, line numbers, and recommendations.

**HTML format:**
```bash
./sidekick scan examples/vulnerable_code.go --format html -o report.html
```
Generates HTML report with all findings properly formatted.

## Migration Notes

- The `Issues []SecurityIssue` field is kept for backward compatibility but is deprecated
- All new code should use `RawFindings` and `HasIssues` fields
- The 3-stage pipeline (Context Analysis → Targeted Scan → Format) is still in place
- Only the final stage (formatting) was simplified to avoid data loss
