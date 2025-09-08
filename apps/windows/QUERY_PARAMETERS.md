# SnipQ Query Parameters Guide

SnipQ supports powerful query parameters that allow you to customize snippet output on-the-fly, similar to URL query strings.

## Basic Syntax

```
:trigger?param1=value1&param2=value2
```

## Parameter Priority

When parameters are specified in multiple places, they follow this priority order:

1. **Query Parameters** (highest priority)
2. **Snippet Defaults**
3. **Global Defaults** (lowest priority)

## Example Snippets with Query Parameters

### 1. Multilingual Thanks (`:ty`)

**Default behavior:**
```
:ty → "Thank you."
```

**With query parameters:**
```
:ty?lang=vi → "Cảm ơn bạn."
:ty?lang=vi&tone=casual → "Cảm ơn bạn nha!"
:ty?lang=ja → "ありがとうございます。"
:ty?lang=ja&tone=casual → "ありがとう！"
:ty?lang=es&tone=formal → "Muchas gracias."
:ty?tone=casual → "Thanks!"
:ty?tone=formal → "Thank you very much."
```

**Available parameters:**
- `lang`: `en` (default), `vi`, `ja`, `es`
- `tone`: `neutral` (default), `casual`, `formal`

### 2. Date Formatting (`:date`)

**Default behavior:**
```
:date → "2025-09-08"
```

**With query parameters:**
```
:date?format=Monday, January 2, 2006 → "Sunday, September 8, 2025"
:date?format=02/01/2006 → "08/09/2025"
:date?format=2006-01-02T15:04:05Z07:00 → "2025-09-08T14:30:45+07:00"
:date?tz=UTC → "2025-09-08" (in UTC)
:date?format=15:04&tz=UTC → "07:30" (time only in UTC)
```

**Available parameters:**
- `format`: Any Go time format string (default: `2006-01-02`)
- `tz`: `Local` (default), `UTC`, or any IANA timezone

### 3. UUID Generation (`:uuid`)

**Default behavior:**
```
:uuid → "550e8400-e29b-41d4-a716-446655440000"
```

**With query parameters:**
```
:uuid?hyphens=false → "550e8400e29b41d4a716446655440000"
:uuid?upper=true → "550E8400-E29B-41D4-A716-446655440000"
:uuid?hyphens=false&upper=true → "550E8400E29B41D4A716446655440000"
```

**Available parameters:**
- `hyphens`: `true` (default), `false`
- `upper`: `false` (default), `true`

### 4. Email Templates (`:email`)

**Default behavior:**
```
:email → "Dear there,\n\nI'm writing to...\n\nBest regards,\n[Your Name]"
```

**With query parameters:**
```
:email?name=John → "Dear John,..."
:email?name=John&greeting=casual → "Hi John,..."
:email?name=Mr. Smith&lang=vi → "Kính chào Mr. Smith,... (Vietnamese)"
:email?greeting=formal&lang=vi → "Kính chào there,... (Vietnamese formal)"
```

**Available parameters:**
- `name`: recipient name (default: `"there"`)
- `greeting`: `formal` (default), `casual`
- `lang`: `en` (default), `vi`

### 5. Meeting Notes (`:meeting`)

**Default behavior:**
```
:meeting → Full meeting template with agenda, discussion, action items
```

**With query parameters:**
```
:meeting?title=Sprint Planning → "# Sprint Planning - September 8, 2025"
:meeting?title=1:1 Review&format=simple → Simplified template without detailed sections
:meeting?attendees=John, Jane, Bob → Updates attendees list
```

**Available parameters:**
- `title`: meeting title (default: `"Team Meeting"`)
- `attendees`: attendees list (default: `"Team Members"`)
- `format`: `detailed` (default), `simple`

### 6. Timestamps (`:time`)

**Default behavior:**
```
:time → "2025-09-08T14:30:45+07:00" (ISO format)
```

**With query parameters:**
```
:time?type=unix → "1725780645"
:time?type=readable → "Sunday, September 8, 2025 at 2:30 PM"
:time?type=short → "09/08/2025 14:30"
:time?type=short&tz=UTC → "09/08/2025 07:30" (UTC time)
```

**Available parameters:**
- `type`: `iso` (default), `unix`, `readable`, `short`, or custom format
- `tz`: `Local` (default), `UTC`, or any IANA timezone

## Template Functions

Query parameters work with these built-in template functions:

- `date "format" "timezone"` - Format current date/time
- `uuid boolean` - Generate UUID (with/without hyphens)
- `upper string` - Convert to uppercase
- `lower string` - Convert to lowercase
- `eq a b` - Check equality
- `ne a b` - Check inequality
- `random` - Generate random values

## Parameter Type Conversion

Query parameter values are automatically converted:

- `"true"`, `"1"`, `"yes"`, `"on"` → `true` (boolean)
- `"false"`, `"0"`, `"no"`, `"off"` → `false` (boolean)
- Everything else remains as string

## Creating Your Own Parametrized Snippets

1. **Define defaults** in your snippet YAML:
```yaml
defaults:
  lang: "en"
  tone: "neutral"
  format: "simple"
```

2. **Use parameters in templates**:
```yaml
template: |
  {{ if eq .lang "vi" }}
    Xin chào {{ .name }}!
  {{ else }}
    Hello {{ .name }}!
  {{ end }}
```

3. **Test with query parameters**:
```
:greeting?name=John&lang=vi
```

## Tips

1. **Parameter names are case-sensitive**
2. **Use URL encoding** for special characters in values
3. **Combine multiple parameters** with `&`
4. **Override any default** with query parameters
5. **Check the "Used Parameters" section** in the test interface to see what values were actually used

## Advanced Examples

**Complex date formatting:**
```
:date?format=Monday, January 2, 2006 at 3:04 PM MST&tz=America/New_York
```

**Multilingual email with custom greeting:**
```
:email?name=Nguyễn Văn A&lang=vi&greeting=formal
```

**Meeting notes for specific team:**
```
:meeting?title=Backend Team Standup&attendees=Alice (Tech Lead), Bob (Senior), Carol (Junior)&format=simple
```
