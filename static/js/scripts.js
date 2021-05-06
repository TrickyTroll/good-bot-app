function playCast(currentDiv) {
    x = currentDiv.parentElement
    var audio = x.parentElement.querySelector('audio');
	var cast = x.parentElement.querySelector('asciinema-player');
    var playPause = audio.getAttribute('class');

	cast.addEventListener('play', function(e) {
        audio.play();
        console.log("Audio playing");
    })

	cast.addEventListener('pause', function(e) {
        audio.pause();
        console.log("Audio paused");
    })

    if (playPause === 'play') {
		cast.play();
        audio.play();
        audio.setAttribute('class', 'pause');

    } else {
        cast.pause();
        audio.play();
        audio.setAttribute('class', 'play');
    }

    audio.onended = function () {
        cast.setAttribute('class', 'play');
    }
}
