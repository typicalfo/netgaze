# Development Guidelines

This document outlines important development requirements and guidelines for the netgaze project.

## No Special Characters Requirement

**Critical Requirement**: Do not use emoji, unicode symbols, or special characters in any code, templates, documentation, or output.

### Why This Matters
- Users will frequently copy/paste output from netgaze
- Special characters (✅, ❌, ⚠️, etc.) complicate downstream processing
- Plain text ensures maximum compatibility with scripts, logs, and analysis tools
- Maintains professional appearance in all terminal environments

### What to Avoid
- Emoji: ✅, ❌, ⚠️, 🚀, 📊, 🎯, 🔍, etc.
- Unicode symbols: ✓, ✗, →, ←, ↑, ↓, etc.
- Special characters: bullets, arrows, decorative symbols
- ANSI escape sequences beyond basic colors

### What to Use Instead
- Status indicators: "Success", "Error", "Warning", "Failed"
- Navigation: "Use arrow keys", "Press 1-3 for tabs"
- Lists: Plain text with hyphens or numbers
- Separators: Simple dashes or pipes

### Examples

**Instead of:**
```
✅ DNS: Success
❌ Ping: Failed
⚠️ TLS: Expired
```

**Use:**
```
DNS: Success
Ping: Failed
TLS: Expired
```

**Instead of:**
```
→ Next step
← Previous
↑ Scroll up
↓ Scroll down
```

**Use:**
```
Next step
Previous
Scroll up
Scroll down
```

## Implementation Checklist

When implementing any component, verify:

- [ ] No emoji in code comments or documentation
- [ ] No unicode symbols in user-facing text
- [ ] Templates use plain text status indicators
- [ ] Error messages are plain text
- [ ] Help text uses standard ASCII characters
- [ ] Log output contains only printable ASCII
- [ ] Configuration files use plain text keys/values

## Template Guidelines

All templates (text, markdown, raw) must:
- Use "Success"/"Failed"/"Warning" instead of symbols
- Use standard ASCII punctuation
- Avoid decorative unicode characters
- Ensure copy-paste compatibility

## TUI Guidelines

The terminal UI should:
- Use Lip Gloss colors for visual hierarchy
- Use plain text for all labels and status
- Avoid unicode box drawing characters (use standard borders)
- Ensure all text is selectable and copyable

## CLI Output Guidelines

Command-line output must:
- Use standard ASCII for all messages
- Avoid decorative symbols in progress indicators
- Use simple text for status updates
- Ensure pipe compatibility

This requirement ensures netgaze output is universally compatible and professional while maintaining excellent user experience.