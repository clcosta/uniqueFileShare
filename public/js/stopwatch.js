class Stopwatch {

    running = false;

    constructor(timerId) {
        this.timerId = timerId;
    }

    updateTimer(endTimeInSeconds) {
        const timerElement = document.getElementById(this.timerId);
        const minute = Math.floor((endTimeInSeconds % 3600) / 60);
        const seconds = (endTimeInSeconds % 3600) % 60;
        timerElement.innerText = `${minute} min: ${seconds} seg`;
        timerElement.classList.add("text-green-600");
        if (endTimeInSeconds <= 0) {
            timerElement.innerText = "Link Expired!";
            timerElement.classList.remove("text-green-600");
            timerElement.classList.add("text-red-500");
            return;
        }
        setTimeout(
            () => {
                if (this.running) {
                    this.updateTimer(endTimeInSeconds - 1);
                }
            },
            1000
        );
    }

    async _getStats(link) {
        const response = await fetch("/stats/" + link);
        if (response.ok) return await response.json();
        return null;
    }

    async checkLinkStats() {
        const url = window.location.href;
        const link = url.split("/")[url.split("/").length - 1];
        const stats = await this._getStats(link);
        if (!stats) {
            this.running = false;
            this.updateTimer(0);
        } else if (stats.Downloaded || stats.ExpiresInSeconds <= 0) {
            this.running = false;
            this.updateTimer(0);
        } else {
            this.running = true;
            this.updateTimer(stats.ExpiresInSeconds);
        }
    }


}

document.addEventListener("DOMContentLoaded", async () => {
    const stopwatch = new Stopwatch("stopwatch");
    await stopwatch.checkLinkStats();

    const linkEl = document.getElementById("link");
    linkEl.addEventListener('auxclick', async () => {
        await stopwatch.checkLinkStats();
    })
});
