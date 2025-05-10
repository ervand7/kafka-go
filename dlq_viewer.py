#!/usr/bin/env python3
"""
DLQ Viewer / Fixer
------------------
Reads messages from the `orders-dlq` topic, shows the error envelope,
and lets you re-publish the original payload back to the `orders` topic.

Envelope format produced by the processor:

{
  "source_topic": "orders",
  "error":        "json_unmarshal",
  "payload":      "<base64 raw bytes>",
  "ts":           1715200000000
}
"""

import base64
import json
import sys
from kafka import KafkaConsumer, KafkaProducer

DLQ_TOPIC       = "orders-dlq"
ORIGINAL_TOPIC  = "orders"

consumer = KafkaConsumer(
    DLQ_TOPIC,
    bootstrap_servers="kafka:9092",
    group_id="dlq-viewer",
    auto_offset_reset="earliest",
    enable_auto_commit=False,
    value_deserializer=lambda v: v,   # keep raw bytes
)

producer = KafkaProducer(
    bootstrap_servers="kafka:9092",
    key_serializer=lambda k: k.encode() if k else None,
    value_serializer=lambda v: v,
)

def pretty_print(env: dict):
    print("\n🟥  DLQ RECORD ------------------------------------")
    print(f"  error : {env.get('error')}")
    print(f"  source: {env.get('source_topic')}")
    print(f"  ts    : {env.get('ts')}")
    print("  payload (decoded):")
    try:
        decoded = base64.b64decode(env["payload"])
        # attempt pretty JSON
        try:
            print(json.dumps(json.loads(decoded), indent=4))
        except Exception:
            print(decoded)
    except KeyError:
        print("  <no payload>")
    print("---------------------------------------------------")

def main() -> None:
    print("👀  DLQ viewer started. Press Ctrl-C to quit.")
    try:
        for msg in consumer:
            try:
                envelope = json.loads(msg.value.decode())
            except Exception as e:
                print("\n⚠️  Could not parse envelope:", e)
                print("Raw:", msg.value[:200], "...")
                consumer.commit()
                continue

            pretty_print(envelope)

            # Ask operator
            ans = input("↩  Type 'r' + ENTER to retry, any other key to skip: ")
            if ans.strip().lower() == "r":
                try:
                    payload_raw = base64.b64decode(envelope["payload"])
                    producer.send(
                        ORIGINAL_TOPIC,
                        key=b"retry",
                        value=payload_raw,
                    )
                    producer.flush()
                    print("✅  Re-published to", ORIGINAL_TOPIC)
                except Exception as e:
                    print("❌  Re-publish failed:", e)
            else:
                print("⏭  Skipped.")

            consumer.commit()
    except KeyboardInterrupt:
        print("\n👋  Exiting viewer.")
    finally:
        consumer.close()
        producer.close()

if __name__ == "__main__":
    main()
