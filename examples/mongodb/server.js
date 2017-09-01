const express = require('express')
const app = express()
const bodyParser = require('body-parser')
const MongoClient = require('mongodb').MongoClient

var db

MongoClient.connect('mongodb://localhost:27017/quotations', (err, database) => {
  if (err) return console.log(err)
  db = database
  app.listen(process.env.PORT || 3000, () => {
    console.log('listening on 3000')
  })
})

app.set('view engine', 'ejs')
app.use(bodyParser.urlencoded({extended: true}))
app.use(bodyParser.json())
app.use(express.static('public'))

app.get('/', (req, res) => {
  db.collection('quotes').find().toArray((err, result) => {
    if (err) return console.log(err)
    users = db.collection('users').find().toArray((err, r) => {
    res.render('index.ejs', {quotes: result, nested: false, users: r, findAuthor: findAuthor})
    })
  })
})

app.get('/nested', (req, res) => {
  db.collection('nested_quotes').find().toArray((err, result) => {
    if (err) return console.log(err)
    res.render('index.ejs', {quotes: result, nested: true})
  })
})

var findAuthor = function(quote, authors) {
  for(i=0; i<authors.length; i++){
    if (authors[i]._id === quote.author) {
      return authors[i]
    }
  }
  return nil
}
