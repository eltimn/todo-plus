
function submitListener(formSelector, callback) {
  const $form = document.querySelector(formSelector)

  $form.addEventListener('submit', async event => {
    event.preventDefault()

    try {
      const formData = new FormData($form)
      const searchParams = new URLSearchParams(formData)

      const req = new Request($form.getAttribute('action'), {
        redirect: "manual"
      })

      const resp = await fetch(req, {
        method: $form.getAttribute('method'),
        body: searchParams
      })

      console.info("resp:", resp)

      // if it's a bad response, try to extract the error from the returned json
      if (!resp.ok) {
        const respJson = await resp.json()
        console.info("respJson:", respJson)
        if ('error' in respJson) {
          callback({
            ok: false,
            error: respJson.error
          })
        } else {
          callback({
            ok: false,
            error: 'unknown error'
          })
        }
      } else {
        callback({ok: true})
      }
    } catch(err) {
      callback({
        ok: false,
        error: err
      })
    }
  })
}

export { submitListener }
