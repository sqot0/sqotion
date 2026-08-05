import mimetypes
import os

import boto3
import httpx


def _s3_client():
    return boto3.client(
        "s3",
        endpoint_url=os.getenv("S3_ENDPOINT"),
        aws_access_key_id=os.getenv("S3_ACCESS_KEY"),
        aws_secret_access_key=os.getenv("S3_SECRET_KEY"),
        region_name=os.getenv("S3_REGION", "auto"),
    )


def save_note(
    bucket: str,
    prefix: str,
    relative_path: str,
    frontmatter: dict,
    markdown_body: str,
    image_urls: list[str],
    file_ids: list[str],
) -> str:
    import yaml

    prefix_dir = prefix.strip("/") + "/" if prefix.strip("/") else ""

    # Build YAML front matter
    yaml_block = (
        "---\n"
        + yaml.dump(
            frontmatter, allow_unicode=True, default_flow_style=False, sort_keys=False
        ).strip()
        + "\n---"
    )

    # Download each image and upload to S3
    image_refs = []
    s3 = _s3_client()

    for i, (url, fid) in enumerate(zip(image_urls, file_ids)):
        try:
            resp = httpx.get(url, timeout=30)
            resp.raise_for_status()

            ext = _guess_ext(url, resp.headers.get("content-type", ""))
            filename = f"{fid}.{ext.split('/')[-1]}"

            img_key = f"{prefix_dir}assets/{filename}"
            s3.put_object(Bucket=bucket, Key=img_key, Body=resp.content)
            print(f"  Uploaded attachment: s3://{bucket}/{img_key}", flush=True)

            image_refs.append(f"![Image {i + 1}]({f"assets/{filename}"})")
        except Exception as e:
            print(f"  Failed to upload attachment from {url}: {e}", flush=True)

    sources = ""
    if image_refs:
        sources = "\n\n## Źródła\n\n" + "\n\n".join(image_refs)

    full_content = yaml_block + "\n" + markdown_body + sources

    md_key = f"{prefix_dir}{relative_path}.md"
    s3.put_object(
        Bucket=bucket,
        Key=md_key,
        Body=full_content.encode("utf-8"),
        ContentType="text/markdown",
    )
    print(f"  Uploaded note: s3://{bucket}/{md_key}", flush=True)
    return md_key


def _guess_ext(url: str, content_type: str) -> str:
    if content_type.startswith("image/"):
        ext = mimetypes.guess_extension(content_type)
        if ext:
            return ext
    ext, _ = mimetypes.guess_type(url)
    if ext:
        return ext
    return "jpg"
