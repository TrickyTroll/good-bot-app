import os
import yaml
import docker
from pathlib import Path

DATA_PATH = Path("../data")

if not DATA_PATH.exists():
    os.mkdir(DATA_PATH)

def project_already_exists(user_id: str, title: str) -> bool:
    """
    project_already_exists helps in making sure that a project does not
    already exist before saving.

    Args:
        user_id (str): The user's id. Should come from the variable
        current_user.id.
        title (str): The project's title. Comes from the form posted by
        the user.

    Returns:
        bool: Whether or not the project already exists.
    """
    save_path: Path = DATA_PATH / user_id / title

    return save_path.exists()

def save_instructions(user_id: str, title: str, contents: str) -> Path:
    """
    save_instructions saves a POST request that contains a script from the
    user to a new file under the user's data directory.

    The path towards the users data directory should be:
        /data/<user_id>

    Args:
        user_id (str): The user's id. Should come from the variable
        current_user.id.
        title (str): The project's title. Comes from the form posted
        by the user.
        contents (str): The contents of the script. Should also come
        from the form posted by the user.

    Returns:
        Path: The path towards the newly saved script
    """
    save_path: Path = (DATA_PATH / user_id / title).with_suffix(".yaml")
    with open(save_path, "w") as stream:
        yaml.safe_dump(contents, stream)

    return save_path

def record_video(instructions_path: Path) -> Path:
    """
    record_video records a video from the provided instructions. It uses
    Good Bot's docker image to do so.

    Args:
        instructions_path (Path): The path towards the instructions
        file. The parent of this path is also where the project's
        directory will be saved.#!/usr/bin/env python
    Returns:
        Path: The path towards the newly created project.
    """
    title: str = instructions_path.stem
    client: docker.DockerClient = docker.from_env()
    image_name: str = "trickytroll/good-bot"
    all_containers: list = client.containers.list()
    should_pull: bool = True
    for container in all_containers:
        if image_name in container.name:
            should_pull = False
    if should_pull:
        client.images.pull(image_name)

    client.containers.run(
        "trickytroll/good-bot",
        command = f"setup --project-path /project/{instructions_path.stem} /project/{instructions_path.name}",
        environment = ["GOOGLE_APPLICATION_CREDENTIALS='/credentials/good-bot-tts.json'"],
        mounts = [
            docker.Mount("/project", instructions_path.parent)
        ]
    )

    client.containers.run(
        "trickytroll/good-bot",
        command = f"record /project/{instructions_path.name}",
        environment = ["GOOGLE_APPLICATION_CREDENTIALS='/credentials/good-bot-tts.json'"],
        mounts = [
            docker.Mount("/credentials", "/home/tricky/Documents/credentials"),
            docker.Mount("/project", instructions_path.parent)
        ]
    )
    return instructions_path.parent / instructions_path.stem
