from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from app.schemas.npc import (
    Perception,
    AgentAction,
    GroupDecideRequest,
    GroupDecideResponse,
)
from app.schemas.world import WorldSnapshot, WorldEvent
from app.graphs.npc_graph import compiled_graph
from app.graphs.group_graph import compiled_group_graph
from app.graphs.world_graph import decide_world_event

app = FastAPI()


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    print("--- VALIDATION ERROR (422) ---")
    print(exc.errors())
    print(exc.body)
    return JSONResponse(
        status_code=422, content={"detail": exc.errors(), "body": exc.body}
    )


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/decide", response_model=AgentAction)
async def decide(perception: Perception) -> AgentAction:
    try:
        result = await compiled_graph.ainvoke(
            {"perception": perception, "messages": [], "action": None}
        )
        return result["action"]
    except Exception as exc:
        print(f"[AGENT FALLBACK] {perception.name}: {exc}")
        return AgentAction(
            internal_thought="My decision process was unavailable, so I will wait.",
            action="WAIT",
        )


@app.post("/decide-batch", response_model=GroupDecideResponse)
async def decide_batch(req: GroupDecideRequest) -> GroupDecideResponse:
    result = await compiled_group_graph.ainvoke(
        {"queue": req.members, "scene_log": [], "actions": {}}
    )
    return GroupDecideResponse(actions=result["actions"])


@app.post("/world-decide", response_model=WorldEvent)
def world_decide(snapshot: WorldSnapshot) -> WorldEvent:
    return decide_world_event(snapshot)
