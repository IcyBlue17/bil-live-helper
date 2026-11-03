// Fuzzy area search across all groups, matching Chinese names, full pinyin
// ("changjian") and initials ("cj").
import Fuse from "fuse.js";
import { pinyin } from "pinyin-pro";
import type { Area, AreaGroup } from "../mygo";

export interface AreaHit {
  group: AreaGroup;
  area: Area;
}

interface Entry extends AreaHit {
  name: string;
  full: string;
  initials: string;
  groupName: string;
}

const py = (s: string) => pinyin(s, { toneType: "none", type: "array", nonZh: "consecutive" }).join("").toLowerCase();
const initials = (s: string) => pinyin(s, { pattern: "first", toneType: "none", type: "array", nonZh: "consecutive" }).join("").toLowerCase();

export function areaIndex(groups: AreaGroup[]) {
  const entries: Entry[] = groups.flatMap((group) =>
    group.areas.map((area) => ({ group, area, name: area.name, full: py(area.name), initials: initials(area.name), groupName: group.name })),
  );
  const fuse = new Fuse(entries, {
    keys: [
      { name: "name", weight: 3 },
      { name: "initials", weight: 2 },
      { name: "full", weight: 2 },
      { name: "groupName", weight: 1 },
    ],
    threshold: 0.34,
    ignoreLocation: true,
  });
  return (q: string): AreaHit[] => fuse.search(q.trim().toLowerCase(), { limit: 60 }).map((r) => r.item);
}

export function findArea(groups: AreaGroup[], id: number): AreaHit | null {
  for (const group of groups) for (const area of group.areas) if (area.id === id) return { group, area };
  return null;
}
