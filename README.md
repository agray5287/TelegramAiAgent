# TelegramAiAgent

Architecture

graph TD
A[Telegram User] --> B(Telegram Bot API)
B --> C(Bot Backend Service)
C --> D[Conversation DB SQLite]
C --> E[OpenAI GPT-4 API]
C --> F[Scheduler Cron]
C --> G[Google Calendar API]:::optional
C --> H[Home Assistant API]:::optional
E --> C
F --> C
G --> C
H --> C
C --> B
B --> A

classDef optional fill:#f2f2f2,stroke:#999,stroke-dasharray: 5 5;
