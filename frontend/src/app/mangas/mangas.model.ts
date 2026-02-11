export interface Manga {
  manga_id: number;
  name: string;
  slug: string;
  status: string;
  provider: string;
  url: string;
  cover_path?: string;
  manga_path: string;

  localized_name?: string;
  publication_status?: string;
  summary?: string;
  start_year?: number;
  start_month?: number;
  start_day?: number;
  author?: string;
  art?: string;
  web_link?: string;
  metadata_updated_at?: string;
  created_at?: string;
}