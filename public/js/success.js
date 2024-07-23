const url = window.location.href;
const linkEl = document.getElementById('link');
linkEl.href = url.replace('success', 'd');
linkEl.innerText = linkEl.href;

function copyLink() {
    navigator.clipboard.writeText(linkEl.href).then(() => {
        const button = document.getElementById('copy-button');
        button.innerHTML = `
            <div class="flex items-center justify-center space-x-2">
                <span>Copied to clipboard</span>
                <img src="/static/images/check.png" alt="check icon" width="24px" height="24px">
            </div>
        `;
        button.classList.remove('bg-red-600', 'hover:bg-red-700');
        button.classList.add('bg-green-600', 'hover:bg-green-700');
    }).catch(err => {
        console.error('Error copying text: ', err);
    });
}

document.addEventListener('DOMContentLoaded', () => {
    const button = document.getElementById('copy-button');
    button.addEventListener('click', copyLink);
})