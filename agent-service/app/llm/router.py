import asyncio
import logging
import re
import time
from collections import defaultdict
from dotenv import load_dotenv
from langchain_core.globals import set_debug
from langchain_cerebras import ChatCerebras
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_groq import ChatGroq
from langchain_mistralai import ChatMistralAI
from langchain_ollama import ChatOllama
from app.schemas.npc import AgentAction
from app.prompts.npc_prompt import PROMPT

load_dotenv()
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - [%(levelname)s] - %(message)s"
)
logger = logging.getLogger(__name__)
set_debug(False)
MAX_CONCURRENT_LLM_CALLS = 3
PROVIDER_TIMEOUT_SECONDS = 12
_llm_semaphore = asyncio.Semaphore(MAX_CONCURRENT_LLM_CALLS)
_provider_cooldown_until: dict[str, float] = defaultdict(float)
_provider_failure_counts: dict[str, int] = defaultdict(int)
groq = ChatGroq(
    model="openai/gpt-oss-20b",
    temperature=0.7,
    max_retries=0,
    timeout=PROVIDER_TIMEOUT_SECONDS,
)
gemini = ChatGoogleGenerativeAI(
    model="gemini-3.5-flash",
    temperature=0.7,
    max_retries=0,
    timeout=PROVIDER_TIMEOUT_SECONDS,
)
cerebras = ChatCerebras(
    model="gemma-4-31b",
    temperature=0.7,
    max_retries=0,
    timeout=PROVIDER_TIMEOUT_SECONDS,
)
mistral = ChatMistralAI(
    model="mistral-small-latest",
    temperature=0.7,
    max_retries=0,
    timeout=PROVIDER_TIMEOUT_SECONDS,
)
OLLAMA_TIMEOUT_SECONDS = 40
ollama_local = ChatOllama(
    model="qwen3:0.6b", temperature=0.7, timeout=OLLAMA_TIMEOUT_SECONDS
).bind(think=False)
PROVIDERS = [
    ("groq", groq, {"reasoning_effort": "low"}),
    ("gemini", gemini, {}),
    ("cerebras", cerebras, {}),
    ("mistral", mistral, {}),
    ("ollama", ollama_local, {}),
]
_PROVIDER_TIMEOUTS: dict[str, int] = {"ollama": OLLAMA_TIMEOUT_SECONDS}


class AllProvidersFailed(RuntimeError):
    pass


def _is_available(name: str) -> bool:
    return time.monotonic() >= _provider_cooldown_until[name]


def _extract_retry_after_seconds(error_text: str) -> int | None:
    match = re.search(
        r"(?:retry in|try again in)\s+(\d+(?:\.\d+)?)s", error_text, re.IGNORECASE
    )
    if match:
        return max(1, int(float(match.group(1))))
    match = re.search(
        r"retry(?:-after)?[^\d]{0,20}(\d+)\s*s", error_text, re.IGNORECASE
    )
    if match:
        return max(1, int(match.group(1)))
    return None


def _cooldown_for_error(provider: str, exc: Exception) -> tuple[int, str]:
    text = str(exc)
    lower = text.lower()
    if "402" in lower or "payment required" in lower:
        return (3600, "billing/quota unavailable")
    if "401" in lower or "403" in lower or "authentication" in lower:
        return (3600, "authentication/permission failure")
    if "429" in lower or "resource_exhausted" in lower or "rate limit" in lower:
        retry_after = _extract_retry_after_seconds(text)
        return (retry_after or 60, "rate limited")
    if any((code in lower for code in ("408", "500", "502", "503", "504"))):
        return (15, "transient provider error")
    if "timeout" in lower or "timed out" in lower or "deadline" in lower:
        return (15, "provider timeout")
    return (10, "provider request failure")


def _record_failure(provider: str, exc: Exception) -> None:
    cooldown_seconds, reason = _cooldown_for_error(provider, exc)
    _provider_failure_counts[provider] += 1
    _provider_cooldown_until[provider] = max(
        _provider_cooldown_until[provider], time.monotonic() + cooldown_seconds
    )
    logger.warning(
        "Provider %s temporarily disabled for %ss (%s). failures=%s",
        provider,
        cooldown_seconds,
        reason,
        _provider_failure_counts[provider],
    )


def _record_success(provider: str) -> None:
    _provider_failure_counts[provider] = 0
    _provider_cooldown_until[provider] = 0.0


async def invoke_agent_model(messages, tools=None):
    async with _llm_semaphore:
        errors: list[str] = []
        for name, provider, bind_kwargs in PROVIDERS:
            if not _is_available(name):
                continue
            try:
                runnable = provider.bind_tools(tools) if tools else provider
                if bind_kwargs:
                    runnable = runnable.bind(**bind_kwargs)
                result = await asyncio.wait_for(
                    runnable.ainvoke(messages),
                    timeout=_PROVIDER_TIMEOUTS.get(name, PROVIDER_TIMEOUT_SECONDS),
                )
                _record_success(name)
                return result
            except asyncio.CancelledError:
                raise
            except Exception as exc:
                _record_failure(name, exc)
                errors.append(f"{name}: {exc}")
        if not errors:
            raise AllProvidersFailed("all providers are currently in cooldown")
        raise AllProvidersFailed("; ".join(errors))


model = groq.with_fallbacks([gemini, cerebras, mistral, ollama_local])
structured_model = model.with_structured_output(AgentAction)
chain = PROMPT | structured_model
logger.info(
    "LLM router initialized: primary=groq, fallbacks=gemini,cerebras,mistral,ollama, concurrency=%s, timeout=%ss",
    MAX_CONCURRENT_LLM_CALLS,
    PROVIDER_TIMEOUT_SECONDS,
)
