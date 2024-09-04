document.addEventListener('DOMContentLoaded', () => {
    const url = window.location.href;
    const button = document.getElementById('download-button');
    button.setAttribute('href', url.replace('download', 'd'));
})