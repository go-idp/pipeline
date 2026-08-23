import { describe, expect, it } from 'vitest';
import { appendEnv, fmtDur, parseYamlLite, runDuration, shortId } from './client';
import type { PipelineRecord } from './api';

describe('parseYamlLite', () => {
  const yaml = `name: CI

stages:
  - name: checkout
    jobs:
      - name: frontend
        steps:
          - name: git clone
            command: git clone https://github.com/go-idp/pipeline $PWD
  - name: build
    mode: serial
    jobs:
      - name: build
        steps:
          - name: go build
            command: go build -o pipeline ./cmd/pipeline
`;

  it('parses stages/jobs/steps with names and commands', () => {
    const stages = parseYamlLite(yaml);
    expect(stages).toHaveLength(2);
    expect(stages[0].name).toBe('checkout');
    expect(stages[0].mode).toBe('parallel');
    expect(stages[0].jobs[0].name).toBe('frontend');
    expect(stages[0].jobs[0].steps[0].name).toBe('git clone');
    expect(stages[0].jobs[0].steps[0].cmd).toContain('git clone');
    expect(stages[1].mode).toBe('serial');
    expect(stages[1].jobs[0].steps[0].cmd).toBe('go build -o pipeline ./cmd/pipeline');
  });

  it('ignores comments and empty lines', () => {
    const s = parseYamlLite('# comment\n\nname: x\n\nstages:\n  - name: a\n    jobs:\n      - name: j\n        steps:\n          - name: s\n            command: echo hi');
    expect(s).toHaveLength(1);
    expect(s[0].name).toBe('a');
  });

  it('returns empty for empty input', () => {
    expect(parseYamlLite('')).toEqual([]);
    expect(parseYamlLite('   \n  # only comment\n')).toEqual([]);
  });
});

describe('appendEnv', () => {
  const yaml = 'name: CI\n\nstages:\n  - name: build\n    jobs:\n      - name: j\n        steps:\n          - name: s\n            command: echo hi\n';

  it('returns yaml unchanged when env text is empty', () => {
    expect(appendEnv(yaml, '')).toBe(yaml);
    expect(appendEnv(yaml, '  \n# comment\n')).toBe(yaml);
  });

  it('injects environment block after the name line', () => {
    const out = appendEnv(yaml, 'BUILD_ID=10086\nGITHUB_REF_NAME=main');
    expect(out).toContain('name: CI\nenvironment:\n  BUILD_ID: "10086"\n  GITHUB_REF_NAME: "main"\n');
    expect(out).toContain('stages:');
  });

  it('keeps original stages intact', () => {
    const out = appendEnv(yaml, 'KEY=value');
    expect(out).toContain('stages:\n  - name: build');
  });
});

describe('fmtDur', () => {
  it('formats seconds', () => {
    expect(fmtDur(5)).toBe('5s');
    expect(fmtDur(75)).toBe('1m 15s');
    expect(fmtDur(3725)).toBe('1h 2m');
  });

  it('handles null/undefined', () => {
    expect(fmtDur(null)).toBe('—');
    expect(fmtDur(undefined)).toBe('—');
  });
});

describe('runDuration', () => {
  it('computes duration from timestamps', () => {
    const start = new Date(Date.now() - 100_000).toISOString();
    const end = new Date(Date.now() - 10_000).toISOString();
    const d = runDuration({ id: 'x', name: 'x', status: 'succeeded', started_at: start, succeed_at: end } as PipelineRecord);
    expect(d).toBeGreaterThanOrEqual(88);
    expect(d).toBeLessThanOrEqual(92);
  });

  it('returns null for missing timestamps', () => {
    expect(runDuration({ id: 'x', name: 'x', status: 'succeeded', started_at: '' } as PipelineRecord)).toBeNull();
  });
});

describe('shortId', () => {
  it('truncates long ids to 9 chars', () => {
    expect(shortId('abcdefghijklmn')).toBe('abcdefghi');
    expect(shortId('abc')).toBe('abc');
  });
});
