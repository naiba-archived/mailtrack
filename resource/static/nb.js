$('#switchLanguage').on('change', function () {
    window.location.href = "?locale=" + this.value
});